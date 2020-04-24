package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

// swagger:route GET /api/users Users getUsers
// Retrieves all the users currently known to kVDI.
// responses:
//   200: usersResponse
//   400: error
//   403: error
//   500: error
func (d *desktopAPI) GetUsers(w http.ResponseWriter, r *http.Request) {
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	users, err := sess.GetAllUsers()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(users, w)
}

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
//   "500":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetUser(w http.ResponseWriter, r *http.Request) {
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	user, err := sess.GetUser(getUserFromRequest(r))
	if err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(user, w)
}

// A list of users
// swagger:response usersResponse
type swaggerUsersResponse struct {
	// in:body
	Body []types.User
}

// A single user
// swagger:response userResponse
type swaggerUserResponse struct {
	// in:body
	Body types.User
}
