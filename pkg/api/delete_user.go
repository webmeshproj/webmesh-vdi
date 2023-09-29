/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package api

import (
	"net/http"

	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/kvdi/kvdi/pkg/util/errors"
)

// swagger:operation DELETE /api/users/{user} Users deleteUserRequest
// ---
// summary: Delete the specified user.
// parameters:
//   - name: user
//     in: path
//     description: The user to delete
//     type: string
//     required: true
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/boolResponse"
//	"400":
//	  "$ref": "#/responses/error"
//	"403":
//	  "$ref": "#/responses/error"
//	"404":
//	  "$ref": "#/responses/error"
func (d *desktopAPI) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)
	if err := d.auth.DeleteUser(username); err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
