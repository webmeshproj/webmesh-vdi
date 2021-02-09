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
	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

// GetUsers implements AuthProvider and serves a GET /api/users request
func (a *AuthProvider) GetUsers() ([]*types.VDIUser, error) {
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}
	users, err := a.listUsers()
	if err != nil {
		return nil, err
	}
	res := make([]*types.VDIUser, 0)
	for _, user := range users {
		res = append(res, &types.VDIUser{
			Name:  user.Username,
			Roles: apiutil.FilterUserRolesByNames(roles, user.Groups),
		})
	}

	return res, nil
}

// CreateUser implements AuthProvider and serves a POST /api/users request
func (a *AuthProvider) CreateUser(req *types.CreateUserRequest) error {
	passwdHash, err := common.HashPassword(req.Password)
	if err != nil {
		return err
	}
	user := &User{
		Username:     req.Username,
		PasswordHash: passwdHash,
		Groups:       req.Roles,
	}
	return a.createUser(user)
}

// GetUser implements AuthProvider and serves a GET /api/users/{user} request
func (a *AuthProvider) GetUser(username string) (*types.VDIUser, error) {
	user, err := a.getUser(username)
	if err != nil {
		return nil, err
	}

	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}

	return &types.VDIUser{
		Name:  user.Username,
		Roles: apiutil.FilterUserRolesByNames(roles, user.Groups),
	}, nil
}

// UpdateUser implements AuthProvider and serves a PUT /api/users/{user} request
func (a *AuthProvider) UpdateUser(username string, req *types.UpdateUserRequest) error {
	user := &User{Username: username}
	if len(req.Roles) != 0 {
		user.Groups = req.Roles
	}
	if req.Password != "" {
		passwdHash, err := common.HashPassword(req.Password)
		if err != nil {
			return err
		}
		user.PasswordHash = passwdHash
	}
	return a.updateUser(user)
}

// DeleteUser implements AuthProvider and serves a DELETE /api/users/{user} request
func (a *AuthProvider) DeleteUser(username string) error {
	return a.deleteUser(username)
}
