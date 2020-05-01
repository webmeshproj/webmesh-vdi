package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// swagger:operation PUT /api/users/{user} Users putUserRequest
// ---
// summary: Update the specified user.
// description: Only the provided attributes will be updated.
// parameters:
// - name: user
//   in: path
//   description: The user to update
//   type: string
//   required: true
// - in: body
//   name: userDetails
//   description: The user details to update.
//   schema:
//     "$ref": "#/definitions/UpdateUserRequest"
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) PutUser(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)
	req := apiutil.GetRequestObject(r).(*v1alpha1.UpdateUserRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	if err := d.auth.UpdateUser(username, req); err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// Implemented by the auth provider

// Request containing updates to a user
// swagger:parameters putUserRequest
type swaggerUpdateUserRequest struct {
	// in:body
	Body v1alpha1.UpdateUserRequest
}
