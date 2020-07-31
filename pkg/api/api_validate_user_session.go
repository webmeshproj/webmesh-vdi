package api

import (
	"net/http"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// ValidateUserSession retrieves the JWT token from the X-Session-Token and
// verifies that it is valid.
func (d *desktopAPI) ValidateUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the auth token
		var authToken string

		if authToken = r.Header.Get(TokenHeader); authToken == "" {
			// the websocket route does not receive request headers from noVNC, so the token is passed
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

		// retrieve the jwt secret
		jwtSecret, err := d.secrets.ReadSecret(v1.JWTSecretKey, true)
		if err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}

		// verify the token and retrieve the claims
		session, err := apiutil.DecodeAndVerifyJWT(jwtSecret, authToken)
		if err != nil {
			apiutil.ReturnAPIForbidden(nil, err.Error(), w)
			return
		}

		// let requests to authorize a token with mfa to go through
		if !session.Authorized && apiutil.GetGorillaPath(r) != "/api/authorize" && r.Method != http.MethodPost {
			apiutil.ReturnAPIForbidden(nil, "User session is not authorized", w)
			return
		}

		// Set the request user object with a pointer to the decoded user session
		apiutil.SetRequestUserSession(r, session)

		// serve the next handler
		next.ServeHTTP(w, r)
	})
}
