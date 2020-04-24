package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

// PutUserRequest requests updates to an existing user
type PutUserRequest struct {
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

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
//     "$ref": "#/definitions/PutUserRequest"
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "500":
//     "$ref": "#/responses/error"
func (d *desktopAPI) UpdateUser(w http.ResponseWriter, r *http.Request) {
	req := GetRequestObject(r).(*PutUserRequest)
	userName := getUserFromRequest(r)
	user := &types.User{Name: userName}

	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()

	if req.Roles != nil && len(req.Roles) > 0 {
		user.Roles = make([]*types.Role, 0)
		for _, role := range req.Roles {
			user.Roles = append(user.Roles, &types.Role{Name: role})
		}
		if err := sess.UpdateUser(user); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	if req.Password != "" {
		if err := sess.SetUserPassword(user, req.Password); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	apiutil.WriteOK(w)
}

// Request containing updates to a user
// swagger:parameters putUserRequest
type swaggerUpdateUserRequest struct {
	// in:body
	Body PutUserRequest
}
