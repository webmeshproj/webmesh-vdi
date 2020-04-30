package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
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
func (d *desktopAPI) PutUser(w http.ResponseWriter, r *http.Request) {}

// Implemented by the auth provider

// Request containing updates to a user
// swagger:parameters putUserRequest
type swaggerUpdateUserRequest struct {
	// in:body
	Body v1alpha1.UpdateUserRequest
}
