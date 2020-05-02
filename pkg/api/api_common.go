package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// TokenHeader is the HTTP header containing the user's session token
const TokenHeader = "X-Session-Token"

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
func (d *desktopAPI) returnNewJWT(w http.ResponseWriter, user *v1alpha1.VDIUser, authorized bool) {
	// fetch the JWT signing secret
	secret, err := d.secrets.ReadSecret(v1alpha1.JWTSecretKey, true)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// create a new token
	claims, newToken, err := apiutil.GenerateJWT(secret, user, authorized)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// return the token to the user
	apiutil.WriteJSON(&v1alpha1.SessionResponse{
		Token:      newToken,
		ExpiresAt:  claims.ExpiresAt,
		User:       user,
		Authorized: authorized,
	}, w)
}

// Session response
// swagger:response sessionResponse
type swaggerSessionResponse struct {
	// in:body
	Body v1alpha1.SessionResponse
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
