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

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// Authenticate implements AuthProvider and simply checks the provided password
// in the request against the hash in the file.
func (a *AuthProvider) Authenticate(req *v1.LoginRequest) (*v1.AuthResult, error) {

	user := &v1.VDIUser{
		Name:  req.Username,
		Roles: make([]*v1.VDIUserRole, 0),
	}

	localUser, err := a.getUser(req.Username)
	if err != nil {
		return nil, err
	}

	if !localUser.PasswordMatchesHash(req.Password) {
		return nil, errors.New("Invalid credentials")
	}

	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}

	user.Roles = apiutil.FilterUserRolesByNames(roles, localUser.Groups)
	return &v1.AuthResult{User: user}, nil
}
