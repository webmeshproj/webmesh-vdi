package errors

import (
	"fmt"
)

type UserNotFoundError struct {
	errMsg string
}

func (r *UserNotFoundError) Error() string {
	return r.errMsg
}

func NewUserNotFoundError(user string) error {
	return &UserNotFoundError{
		errMsg: fmt.Sprintf("User '%s' not found in the database", user),
	}
}

func IsUserNotFoundError(err error) bool {
	if _, ok := err.(*UserNotFoundError); ok {
		return true
	}
	return false
}

type RoleNotFoundError struct {
	errMsg string
}

func (r *RoleNotFoundError) Error() string {
	return r.errMsg
}

func NewRoleNotFoundError(role string) error {
	return &RoleNotFoundError{
		errMsg: fmt.Sprintf("Role '%s' not found in the database", role),
	}
}

func IsRoleNotFoundError(err error) bool {
	if _, ok := err.(*RoleNotFoundError); ok {
		return true
	}
	return false
}

type UserSessionNotFoundError struct {
	errMsg string
}

func (r *UserSessionNotFoundError) Error() string {
	return r.errMsg
}

func NewUserSessionNotFoundError(id string) error {
	return &UserSessionNotFoundError{
		errMsg: fmt.Sprintf("User session '%s' not found in the database", id),
	}
}

func IsUserSessionNotFoundError(err error) bool {
	if _, ok := err.(*UserSessionNotFoundError); ok {
		return true
	}
	return false
}
