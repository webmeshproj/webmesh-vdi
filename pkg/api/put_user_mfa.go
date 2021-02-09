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

	"github.com/xlzd/gotp"
)

// swagger:operation PUT /api/users/{user}/mfa Users putUserMFARequest
// ---
// summary: Updates MFA configuration for the specified user.
// parameters:
// - name: user
//   in: path
//   description: The user to update
//   type: string
//   required: true
// - in: body
//   name: putUserMFARequest
//   description: The user details to update.
//   schema:
//     "$ref": "#/definitions/UpdateMFARequest"
// responses:
//   "200":
//     "$ref": "#/responses/updateMFAResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) PutUserMFA(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)

	// Only verify user if not using OIDC. We don't have a way to verify the user
	// otherwise. This does leave the door open for someone with access to this endpoint
	// to go rogue and flood the secrets with bad users.
	if !d.vdiCluster.IsUsingOIDCAuth() {
		if _, err := d.auth.GetUser(username); err != nil {
			if errors.IsUserNotFoundError(err) {
				apiutil.ReturnAPINotFound(err, w)
				return
			}
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	req := apiutil.GetRequestObject(r).(*types.UpdateMFARequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	// We are enabling MFA
	if req.Enabled {
		// https://github.com/xlzd/gotp/blob/master/utils.go#L79
		//Only uses uppercase characters and digits
		newSecret := gotp.RandomSecret(32)
		if err := d.mfa.SetUserMFAStatus(username, newSecret, false); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		apiutil.WriteJSON(&types.MFAResponse{
			Enabled:         true,
			Verified:        false,
			ProvisioningURI: gotp.NewDefaultTOTP(newSecret).ProvisioningUri(username, "kVDI"),
		}, w)
		return
	}

	// We are disabling MFA
	if err := d.mfa.DeleteUserSecret(username); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&types.MFAResponse{
		Enabled: false,
	}, w)
}

// Request containing updates to a user
// swagger:parameters putUserMFARequest
type swaggerUpdateMFARequest struct {
	// in:body
	Body types.UpdateMFARequest
}

// Session response
// swagger:response updateMFAResponse
type swaggerUpdateMFAResponse struct {
	// in:body
	Body types.MFAResponse
}
