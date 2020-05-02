package api

import (
	"errors"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
)

// ValidateUserSession retrieves the JWT token from the X-Session-Token and
// verifies that it is valid.
func (d *desktopAPI) ValidateUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the auth token
		authToken := r.Header.Get(TokenHeader)
		if authToken == "" {
			// the websocket route cannot parse request headers, so the token is passed
			// as a query argument. This effectively gives that option to all routes.
			if keys, ok := r.URL.Query()["token"]; ok {
				authToken = keys[0]
			}
		}

		// if we don't have a token we can't proceed
		if authToken == "" {
			apiutil.ReturnAPIForbidden(nil, "No token provided in request", w)
			return
		}

		// parse the token
		parser := &jwt.Parser{UseJSONNumber: true}
		token, err := parser.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Incorrect signing algorithm on token")
			}
			// use cache for the JWT secret, since we use it for every request
			return d.secrets.ReadSecret(v1alpha1.JWTSecretKey, true)
		})

		// check token validity
		if !token.Valid {

			// Just the error conditions we have specific messages for
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					apiutil.ReturnAPIForbidden(nil, "Malformed token provided in request", w)
					return
				} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
					apiutil.ReturnAPIForbidden(nil, "User session has expired", w)
					return
				} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
					apiutil.ReturnAPIForbidden(nil, "User session is not valid yet", w)
					return
				}
			}

			// Unhandled token error - generic message
			apiutil.ReturnAPIForbidden(err, "Provided token is invalid", w)
			return
		}

		// Retrieve the token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// decode the claims into a session object
			session := &v1alpha1.JWTClaims{}
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  session,
			})
			if err != nil {
				apiutil.ReturnAPIError(err, w)
				return
			}
			if err := decoder.Decode(claims); err != nil {
				apiutil.ReturnAPIError(err, w)
				return
			}

			if !session.Authorized && apiutil.GetGorillaPath(r) != "/api/authorize" && r.Method != http.MethodPost {
				apiutil.ReturnAPIForbidden(nil, "User session is not authorized", w)
				return
			}

			// Set the request user object with a pointer to the decoded user session
			apiutil.SetRequestUserSession(r, session)

			// serve the next handler
			next.ServeHTTP(w, r)
			return
		}

		// The claims in the token weren't as expected
		apiutil.ReturnAPIError(errors.New("Could not parse claims from provided token"), w)
	})
}
