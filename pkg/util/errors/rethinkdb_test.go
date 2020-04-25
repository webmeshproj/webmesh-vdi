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

	// UserSessionNotFoundError

	userSessionNotFound := NewUserSessionNotFoundError("fakeSession")
	if userSessionNotFound.Error() != fmt.Sprintf(userSessionNotFoundFormat, "fakeSession") {
		t.Error("Error message for not found user session is malformed")
	}
	if !IsUserSessionNotFoundError(userSessionNotFound) {
		t.Error("Error should be valid UserSessionNotFoundError")
	}
	if IsUserSessionNotFoundError(errors.New("fake error")) {
		t.Error("Generic error should not evaluate to UserSessionNotFoundError")
	}
}
