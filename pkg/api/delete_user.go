package api

import (
	"net/http"
)

// swagger:operation DELETE /api/users/{user} Users deleteUserRequest
// ---
// summary: Delete the specified user.
// parameters:
// - name: user
//   in: path
//   description: The user to delete
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) DeleteUser(w http.ResponseWriter, r *http.Request) {}

// Implemented by the auth provider
