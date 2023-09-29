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

	"github.com/xlzd/gotp"

	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/kvdi/kvdi/pkg/util/errors"
)

// swagger:operation PUT /api/users/{user}/mfa/verify Users putUserMFAVerifyRequest
// ---
// summary: Verifies the MFA setup for the given user.
// parameters:
//   - name: user
//     in: path
//     description: The user to verify MFA for
//     type: string
//     required: true
//   - in: body
//     name: body
//     description: The MFA token generated by the app
//     schema:
//     "$ref": "#/definitions/AuthorizeRequest"
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/verifyUserMFAResponse"
//	"400":
//	  "$ref": "#/responses/error"
//	"403":
//	  "$ref": "#/responses/error"
//	"404":
//	  "$ref": "#/responses/error"
func (d *desktopAPI) PutUserMFAVerify(w http.ResponseWriter, r *http.Request) {
	req := apiutil.GetRequestObject(r).(*types.AuthorizeRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	username := apiutil.GetUserFromRequest(r)
	token := req.OTP

	secret, alreadyVerified, err := d.mfa.GetUserMFAStatus(username)
	if err != nil {
		if !errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIError(err, w)
			return
		}
		apiutil.ReturnAPINotFound(err, w)
		return
	}

	totp := gotp.NewDefaultTOTP(secret)

	if totp.Now() != token {
		// just return an error, if they are already verified we don't want to
		// change that. This way, this route can also be used by a user to simply
		// make sure their MFA still works.
		apiutil.ReturnAPIForbidden(nil, "Invalid MFA Code", w)
		return
	}

	if !alreadyVerified {
		// We can mark the user as verified now
		if err := d.mfa.SetUserMFAStatus(username, secret, true); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	// Return to the user the token was valid
	apiutil.WriteJSON(&types.MFAResponse{
		Enabled:  true,
		Verified: true,
	}, w)
}

// Request containing an OTP token
// swagger:parameters verifyUserMFARequest
type swaggerVerifyUserMFARequest struct {
	// in:body
	Body types.AuthorizeRequest
}

// Response with mfa details for the user
// swagger:response verifyUserMFAResponse
type swaggerVerifyUserMFAResponse struct {
	// in:body
	Body types.MFAResponse
}
