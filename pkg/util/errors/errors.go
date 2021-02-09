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
	goerrors "errors"
	"strings"
)

// New wraps the stdlib errors.New for simplicity when using this package.
func New(msg string) error {
	return goerrors.New(msg)
}

// IsBrokenPipeError returns true if the error is from trying to write to a
// closed connection.
func IsBrokenPipeError(err error) bool {
	return strings.HasSuffix(err.Error(), "broken pipe")
}
