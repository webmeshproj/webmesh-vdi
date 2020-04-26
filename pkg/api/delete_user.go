package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
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
//   "500":
//     "$ref": "#/responses/error"
func (d *desktopAPI) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(r)
	sess, err := d.getDB()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	if err := sess.DeleteUser(&types.User{Name: user}); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
