package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	corev1 "k8s.io/api/core/v1"
)

// TokenHeader is the HTTP header containing the user's access token
const TokenHeader = "X-Session-Token"

// RefreshTokenCookie is the cookie used to store a user's refresh token
const RefreshTokenCookie = "refreshToken"

// swagger:route GET /api/whoami Miscellaneous whoAmI
// Retrieves information about the current user session.
// responses:
//   200: userResponse
//   403: error
//   500: error
func (d *desktopAPI) GetWhoAmI(w http.ResponseWriter, r *http.Request) {
	session := apiutil.GetRequestUserSession(r)
	apiutil.WriteJSON(session.User, w)
}

// returnNewJWT will return a new JSON web token to the requestor.
func (d *desktopAPI) returnNewJWT(w http.ResponseWriter, result *v1.AuthResult, authorized bool, state string) {
	// fetch the JWT signing secret
	secret, err := d.secrets.ReadSecret(v1.JWTSecretKey, true)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// create a new token
	claims, newToken, err := apiutil.GenerateJWT(secret, result, authorized, d.vdiCluster.GetTokenDuration())
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	if authorized && !result.RefreshNotSupported {
		// Generate a refresh token
		refreshToken, err := d.generateRefreshToken(result.User)
		if err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		// Set a Secure, HttpOnly cookie so that it can only be used over HTTPS and not
		// accessed by the browser.
		http.SetCookie(w, &http.Cookie{
			Name:     RefreshTokenCookie,
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   true,
		})
	}

	// return the token to the user
	apiutil.WriteJSON(&v1.SessionResponse{
		Token:      newToken,
		ExpiresAt:  claims.ExpiresAt,
		Renewable:  !result.RefreshNotSupported,
		User:       result.User,
		Authorized: authorized,
		State:      state,
	}, w)
}

func (d *desktopAPI) generateRefreshToken(user *v1.VDIUser) (string, error) {
	refreshToken := uuid.New().String()
	if err := d.secrets.Lock(); err != nil {
		return "", err
	}
	defer d.secrets.Release()
	tokens, err := d.secrets.ReadSecretMap(v1.RefreshTokensSecretKey, true)
	if err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return "", err
		}
		tokens = make(map[string][]byte)
	}
	tokens[refreshToken] = []byte(user.Name)
	return refreshToken, d.secrets.WriteSecretMap(v1.RefreshTokensSecretKey, tokens)
}

func (d *desktopAPI) lookupRefreshToken(refreshToken string) (string, error) {
	if err := d.secrets.Lock(); err != nil {
		return "", err
	}
	defer d.secrets.Release()
	tokens, err := d.secrets.ReadSecretMap(v1.RefreshTokensSecretKey, true)
	if err != nil {
		if errors.IsSecretNotFoundError(err) {
			return "", errors.New("The refresh token does not exist in the secret storage")
		}
		return "", err
	}
	user, ok := tokens[refreshToken]
	if !ok {
		return "", errors.New("The refresh token does not exist in the secret storage")
	}
	delete(tokens, refreshToken)
	return string(user), d.secrets.WriteSecretMap(v1.RefreshTokensSecretKey, tokens)
}

func (d *desktopAPI) getDesktopWebsocketURL(r *http.Request) (*url.URL, error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &corev1.Service{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return nil, err
	}
	return url.Parse(fmt.Sprintf("wss://%s:%d", found.Spec.ClusterIP, v1.WebPort))
}

func (d *desktopAPI) getDesktopWebURL(r *http.Request) (string, error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &corev1.Service{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s:%d", found.Spec.ClusterIP, v1.WebPort), nil
}

// decodeAndVerifyToken verifies the signature on the provided token and returns the
// embedded claims.
func (d *desktopAPI) decodeAndVerifyToken(authToken string) (*v1.JWTClaims, error) {
	// parse the token
	parser := &jwt.Parser{UseJSONNumber: true}
	token, err := parser.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Incorrect signing algorithm on token")
		}
		// use cache for the JWT secret, since we use it for every request
		return d.secrets.ReadSecret(v1.JWTSecretKey, true)
	})
	if err != nil {
		return nil, err
	}

	// check token validity
	if !token.Valid {

		// Just the error conditions we have specific messages for
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("Malformed token provided in the request")
			} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return nil, errors.New("Token has expired")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("Token is not valid yet")
			}
		}

		// Unhandled token error - generic message
		return nil, errors.New("Token is invalid")
	}

	// Retrieve the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// The claims in the token weren't as expected
		return nil, errors.New("Could not coerce token claims to MapClaims")
	}

	// decode the claims into a session object
	session := &v1.JWTClaims{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  session,
	})
	if err != nil {
		return nil, err
	}
	return session, decoder.Decode(claims)
}

// Session response
// swagger:response sessionResponse
type swaggerSessionResponse struct {
	// in:body
	Body v1.SessionResponse
}

// Success response
// swagger:response boolResponse
type swaggerBoolResponse struct {
	// in:body
	Body struct {
		Ok bool `json:"ok"`
	}
}

// A generic error response
// swagger:response error
type swaggerResponseError struct {
	// in:body
	Body errors.APIError
}
