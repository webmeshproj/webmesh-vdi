package api

import (
	"errors"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:route POST /api/users Users postUserRequest
// Create a new user in kVDI.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) PostUsers(w http.ResponseWriter, r *http.Request) {
	req := apiutil.GetRequestObject(r).(*v1.CreateUserRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	if err := d.auth.CreateUser(req); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// Request containing a new user
// swagger:parameters postUserRequest
type swaggerCreateUserRequest struct {
	// in:body
	Body v1.CreateUserRequest
}
