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

package oidc

import (
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// GetUsers should return a list of VDIUsers.
func (a *AuthProvider) GetUsers() ([]*v1.VDIUser, error) {
	return nil, errors.New("Listing users is not supported when using OIDC authentication")
}

// GetUser should retrieve a single VDIUser.
func (a *AuthProvider) GetUser(username string) (*v1.VDIUser, error) {
	return nil, errors.New("Retrieving user information is not supported when using OIDC authentication")
}

// CreateUser should handle any logic required to register a new user in kVDI.
func (a *AuthProvider) CreateUser(*v1.CreateUserRequest) error {
	return errors.New("Creating users is not supported when using OIDC authentication")
}

// UpdateUser should update a VDIUser.
func (a *AuthProvider) UpdateUser(string, *v1.UpdateUserRequest) error {
	return errors.New("Updating users is not supported when using OIDC authentication")
}

// DeleteUser should remove a VDIUser.
func (a *AuthProvider) DeleteUser(string) error {
	return errors.New("Deleting users is not supported when using OIDC authentication")
}
