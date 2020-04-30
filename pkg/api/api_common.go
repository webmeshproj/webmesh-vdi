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
func (d *desktopAPI) WhoAmI(w http.ResponseWriter, r *http.Request) {
	session := apiutil.GetRequestUserSession(r)
	apiutil.WriteJSON(session.User, w)
}

// swagger:route POST /api/login Auth loginRequest
// Retrieves a new JWT token. This route may behave differently depending on the auth provider.
// responses:
//   200: sessionResponse
//   400: error
//   403: error
//   500: error
func loginDoc(w http.ResponseWriter, r *http.Request) {}

// Login request
// swagger:parameters loginRequest
type swaggerLoginRequest struct {
	// in:body
	Body v1alpha1.LoginRequest
}

// Success response
// swagger:response boolResponse
type swaggerBoolResponse struct {
	// in:body
	Body struct {
		Ok bool `json:"ok"`
	}
}

// Session response
// swagger:response sessionResponse
type swaggerSessionResponse struct {
	// in:body
	Body v1alpha1.SessionResponse
}

// A generic error response
// swagger:response error
type swaggerResponseError struct {
	// in:body
	Body errors.APIError
}
