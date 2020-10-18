package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

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
	if err := d.secrets.Lock(10); err != nil {
		return "", err
	}
	defer d.secrets.Release()
	tokens, err := d.secrets.ReadSecretMap(v1.RefreshTokensSecretKey, false)
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
	if err := d.secrets.Lock(10); err != nil {
		return "", err
	}
	defer d.secrets.Release()
	tokens, err := d.secrets.ReadSecretMap(v1.RefreshTokensSecretKey, false)
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
	host, err := d.getDesktopWebHost(r)
	if err != nil {
		return nil, err
	}
	return url.Parse(fmt.Sprintf("wss://%s", host))
}

func (d *desktopAPI) getDesktopWebHost(r *http.Request) (string, error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &corev1.Service{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", found.Spec.ClusterIP, v1.WebPort), nil
}

func (d *desktopAPI) serveHTTPProxy(w http.ResponseWriter, r *http.Request) {
	desktopHost, err := d.getDesktopWebHost(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	// Overwrite the request object host to point to the desktop container
	u := r.URL
	u.Scheme = "https"
	u.Host = desktopHost
	u.Path = path.Clean(u.Path) + "/" // TODO: this method is used for kvdi-proxy requests which expect trailing slashes. come up with cleaner fix.

	// Buld a request from the source
	req, err := http.NewRequest(r.Method, u.String(), bufio.NewReader(r.Body))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// Copy the headers over to the new request
	for hdr, val := range r.Header {
		req.Header.Add(hdr, strings.Join(val, ";"))
	}

	// Build an HTTP client
	clientTLSConfig, err := tlsutil.NewClientTLSConfig()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientTLSConfig,
		},
	}

	// Do the request
	resp, err := httpClient.Do(req)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	defer resp.Body.Close()

	// copy the response from the proxy to the requestor
	for hdr, val := range resp.Header {
		w.Header().Set(hdr, strings.Join(val, ";"))
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		apiLogger.Error(err, "Error copying response body from desktop proxy")
	}

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
