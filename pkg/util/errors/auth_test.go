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
	"errors"
	"fmt"
	"testing"
)

// TestRethinkDBErrors tests the various db errors
func TestRethinkDBErrors(t *testing.T) {

	// UserNotFoundError

	userNotFound := NewUserNotFoundError("fakeUser")
	if userNotFound.Error() != fmt.Sprintf(userNotFoundFormat, "fakeUser") {
		t.Error("Error message for not found user is malformed")
	}
	if !IsUserNotFoundError(userNotFound) {
		t.Error("Error should be valid UserNotFoundError")
	}
	if IsUserNotFoundError(errors.New("fake error")) {
		t.Error("Generic error should not evaluate to UserNotFoundError")
	}

	// RoleNotFoundError

	roleNotFound := NewRoleNotFoundError("fakeRole")
	if roleNotFound.Error() != fmt.Sprintf(roleNotFoundFormat, "fakeRole") {
		t.Error("Error message for not found role is malformed")
	}
	if !IsRoleNotFoundError(roleNotFound) {
		t.Error("Error should be valid RoleNotFoundError")
	}
	if IsRoleNotFoundError(errors.New("fake error")) {
		t.Error("Generic error should not evaluate to RoleNotFoundError")
	}

}
