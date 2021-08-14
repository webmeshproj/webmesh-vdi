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

// swagger:route POST /api/authorize Auth authorizeRequest
// Authorizes a JWT token with a one time password.
// responses:
//   200: sessionResponse
//   400: error
//   403: error
func (d *desktopAPI) PostAuthorize(w http.ResponseWriter, r *http.Request) {
	userSession := apiutil.GetRequestUserSession(r)

	// retrieve the OTP from the request
	req := apiutil.GetRequestObject(r).(*types.AuthorizeRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	secret, verified, err := d.mfa.GetUserMFAStatus(userSession.User.Name)
	if err != nil {
		if !errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIError(err, w)
			return
		}
		// The user does not require MFA - this shouldn't happen but go ahead
		// and send back an authorized token
		d.returnNewJWT(w, &types.AuthResult{
			User:                userSession.User,
			RefreshNotSupported: !userSession.Renewable,
		}, true, req.GetState())
		return
	}

	if !verified {
		// The user has not verified their MFA secret yet.
		// The login attempt should not have required MFA.
		apiutil.ReturnAPIForbidden(nil, "MFA token has not been verified", w)
		return
	}

	totp := gotp.NewDefaultTOTP(secret)

	if totp.Now() != req.GetOTP() {
		apiutil.ReturnAPIForbidden(nil, "Invalid MFA Code", w)
		return
	}

	d.returnNewJWT(w, &types.AuthResult{
		User:                userSession.User,
		RefreshNotSupported: !userSession.Renewable,
	}, true, req.GetState())
}

// Request containing a one-time password.
// swagger:parameters authorizeRequest
type swaggerAuthorizeRequest struct {
	// in:body
	Body types.AuthorizeRequest
}
