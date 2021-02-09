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

package errors

import (
	"fmt"
)

// Formatting strings for rethinkdb errors
const (
	userNotFoundFormat = "User '%s' not found in the cluster"
	roleNotFoundFormat = "Role '%s' not found in the cluster"
)

// UserNotFoundError is an error signaling that the requested user was not found.
type UserNotFoundError struct {
	errMsg string
}

// Error implements the error interface.
func (r *UserNotFoundError) Error() string {
	return r.errMsg
}

// NewUserNotFoundError returns a new UserNotFoundError for the provided username.
func NewUserNotFoundError(user string) error {
	return &UserNotFoundError{
		errMsg: fmt.Sprintf(userNotFoundFormat, user),
	}
}

// IsUserNotFoundError returns true if the given error interface is a UserNotFoundError.
func IsUserNotFoundError(err error) bool {
	if _, ok := err.(*UserNotFoundError); ok {
		return true
	}
	return false
}

// RoleNotFoundError is an error signaling that the requested role was not found.
type RoleNotFoundError struct {
	errMsg string
}

// Error implements the error interface.
func (r *RoleNotFoundError) Error() string {
	return r.errMsg
}

// NewRoleNotFoundError returns a new RoleNotFoundError for the provided role.
func NewRoleNotFoundError(role string) error {
	return &RoleNotFoundError{
		errMsg: fmt.Sprintf(roleNotFoundFormat, role),
	}
}

// IsRoleNotFoundError returns true if the given error interface is a RoleNotFoundError.
func IsRoleNotFoundError(err error) bool {
	if _, ok := err.(*RoleNotFoundError); ok {
		return true
	}
	return false
}
