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

package local

import (
	"errors"

	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
)

// Authenticate implements AuthProvider and simply checks the provided password
// in the request against the hash in the file.
func (a *AuthProvider) Authenticate(req *types.LoginRequest) (*types.AuthResult, error) {
	user := &types.VDIUser{
		Name:  req.Username,
		Roles: make([]*types.VDIUserRole, 0),
	}
	localUser, err := a.getUser(req.Username)
	if err != nil {
		return nil, err
	}
	if !localUser.PasswordMatchesHash(req.Password) {
		return nil, errors.New("invalid credentials")
	}
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}
	user.Roles = apiutil.FilterUserRolesByNames(roles, localUser.Groups)
	return &types.AuthResult{User: user}, nil
}
