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

	"github.com/tinyzimmer/kvdi/pkg/types"
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
	req := apiutil.GetRequestObject(r).(*types.UpdateUserRequest)
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

// Request containing updates to a user
// swagger:parameters putUserRequest
type swaggerUpdateUserRequest struct {
	// in:body
	Body types.UpdateUserRequest
}
