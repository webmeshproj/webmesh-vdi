package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
)

// swagger:route GET /api/users Users getUsers
// Retrieves all the users currently known to kVDI.
// responses:
//   200: usersResponse
//   400: error
//   403: error
func (d *desktopAPI) GetUsers(w http.ResponseWriter, r *http.Request) {}

// Implemented by the auth provider

// swagger:operation GET /api/users/{user} Users getUser
// ---
// summary: Retrieve the specified user.
// description: Details include the roles, grants, namespaces, and template patterns for the user.
// parameters:
// - name: user
//   in: path
//   description: The username to retrieve details about
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/userResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetUser(w http.ResponseWriter, r *http.Request) {}

// Implemented by the auth provider

// A list of users
// swagger:response usersResponse
type swaggerUsersResponse struct {
	// in:body
	Body []v1alpha1.VDIUser
}

// A single user
// swagger:response userResponse
type swaggerUserResponse struct {
	// in:body
	Body v1alpha1.VDIUser
}
