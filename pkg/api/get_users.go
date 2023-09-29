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

	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/kvdi/kvdi/pkg/util/errors"
)

// swagger:route GET /api/users Users getUsers
// Retrieves all the users currently known to kVDI.
// responses:
//
//	200: usersResponse
//	400: error
//	403: error
func (d *desktopAPI) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := d.auth.GetUsers()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	mfaUsers, err := d.mfa.GetMFAUsers()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	for _, user := range users {
		if verified, ok := mfaUsers[user.Name]; ok {
			user.MFA = &types.UserMFAStatus{
				Enabled:  true,
				Verified: verified,
			}
		} else {
			user.MFA = &types.UserMFAStatus{
				Enabled: false,
			}
		}
	}
	apiutil.WriteJSON(users, w)
}

// swagger:operation GET /api/users/{user} Users getUser
// ---
// summary: Retrieve the specified user.
// description: Details include the roles and grants for the user.
// parameters:
//   - name: user
//     in: path
//     description: The username to retrieve details about
//     type: string
//     required: true
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/userResponse"
//	"400":
//	  "$ref": "#/responses/error"
//	"403":
//	  "$ref": "#/responses/error"
//	"404":
//	  "$ref": "#/responses/error"
func (d *desktopAPI) GetUser(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)
	user, err := d.auth.GetUser(username)
	if err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	if _, verified, err := d.mfa.GetUserMFAStatus(username); err != nil {
		if !errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIError(err, w)
			return
		}
		user.MFA = &types.UserMFAStatus{
			Enabled: false,
		}
	} else if err == nil {
		user.MFA = &types.UserMFAStatus{
			Enabled:  true,
			Verified: verified,
		}
	}
	apiutil.WriteJSON(user, w)
}

// A list of users
// swagger:response usersResponse
type swaggerUsersResponse struct {
	// in:body
	Body []types.VDIUser
}

// A single user
// swagger:response userResponse
type swaggerUserResponse struct {
	// in:body
	Body types.VDIUser
}
