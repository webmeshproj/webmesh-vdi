package errors

import (
	"fmt"
)

// Formatting strings for rethinkdb errors
const (
	userNotFoundFormat        = "User '%s' not found in the database"
	roleNotFoundFormat        = "Role '%s' not found in the database"
	userSessionNotFoundFormat = "User session with token '%s' not found in the database"
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

// UserSessionNotFoundError is an error signaling that the provided user session
// token was not found.
type UserSessionNotFoundError struct {
	errMsg string
}

// Error implements the error interface.
func (r *UserSessionNotFoundError) Error() string {
	return r.errMsg
}

// NewUserSessionNotFoundError returns a new UserSessionNotFoundError for the
// given token.
func NewUserSessionNotFoundError(id string) error {
	return &UserSessionNotFoundError{
		errMsg: fmt.Sprintf(userSessionNotFoundFormat, id),
	}
}

// IsUserSessionNotFoundError returns true if the given error interface is a
// UserSessionNotFoundError.
func IsUserSessionNotFoundError(err error) bool {
	if _, ok := err.(*UserSessionNotFoundError); ok {
		return true
	}
	return false
}
