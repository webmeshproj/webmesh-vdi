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

// swagger:route GET /api/refresh_token Auth refreshTokenRequest
// Retrieves a new JWT access token. It uses the HttpOnly cookie included in the request.
// responses:
//
//	200: sessionResponse
//	400: error
//	403: error
//	500: error
func (d *desktopAPI) GetRefreshToken(w http.ResponseWriter, r *http.Request) {

	if d.vdiCluster.IsUsingOIDCAuth() {
		apiutil.ReturnAPIError(errors.New("Token has expired and cannot be refreshed due to OIDC auth"), w)
		return
	}

	refreshToken, err := r.Cookie(RefreshTokenCookie)
	if err != nil {
		apiutil.ReturnAPIForbidden(err, "Could not retrieve a refresh token from the request", w)
		return
	}
	if refreshToken == nil || refreshToken.Value == "" {
		apiutil.ReturnAPIForbidden(nil, "No refresh token was provided in the request", w)
		return
	}

	username, err := d.lookupRefreshToken(refreshToken.Value)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// retrieve the user from the auth provider
	user, err := d.auth.GetUser(username)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// return a new access and refresh token for the user
	// TODO: Use state during a refresh?
	d.returnNewJWT(w, &types.AuthResult{User: user}, true, "")
}
