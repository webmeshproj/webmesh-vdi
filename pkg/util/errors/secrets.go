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

import "fmt"

// The error message format for a SecretNotFoundError
const secretNotFoundFormat = "Secret '%s' could not be found"

// SecretNotFoundError is used to signal from a secrets backend that the requested
// secret does not exist.
type SecretNotFoundError struct {
	errMsg string
}

// Error implements the error interface
func (r *SecretNotFoundError) Error() string {
	return r.errMsg
}

// NewSecretNotFoundError returns a new SecretNotFoundError for the given resource
// name.
func NewSecretNotFoundError(secret string) error {
	return &SecretNotFoundError{
		errMsg: fmt.Sprintf(secretNotFoundFormat, secret),
	}
}

// IsSecretNotFoundError returns true if the given error is a SecretNotFoundError.
func IsSecretNotFoundError(err error) bool {
	if _, ok := err.(*SecretNotFoundError); ok {
		return true
	}
	return false
}
