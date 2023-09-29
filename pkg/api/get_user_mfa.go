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

	"github.com/xlzd/gotp"
)

// swagger:operation GET /api/users/{user}/mfa Users getUserMFARequest
// ---
// summary: Retrieves MFA status for the given user.
// parameters:
//   - name: user
//     in: path
//     description: The user to query
//     type: string
//     required: true
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/getMFAResponse"
//	"400":
//	  "$ref": "#/responses/error"
//	"403":
//	  "$ref": "#/responses/error"
//	"404":
//	  "$ref": "#/responses/error"
func (d *desktopAPI) GetUserMFA(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)

	secret, verified, err := d.mfa.GetUserMFAStatus(username)
	if err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.WriteJSON(&types.MFAResponse{
				Enabled: false,
			}, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&types.MFAResponse{
		Enabled:         true,
		Verified:        verified,
		ProvisioningURI: gotp.NewDefaultTOTP(secret).ProvisioningUri(username, "kVDI"),
	}, w)
}

// Session response
// swagger:response getMFAResponse
type swaggerGetMFAResponse struct {
	// in:body
	Body types.MFAResponse
}
