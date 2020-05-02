package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSecretNotFoundError(t *testing.T) {
	qerr := NewSecretNotFoundError("fake secret")

	if qerr.Error() != fmt.Sprintf(secretNotFoundFormat, "fake secret") {
		t.Error("Error body is malformed")
	}

	if ok := IsSecretNotFoundError(qerr); !ok {
		t.Error("Should be a valid requeue error")
	}

	if ok := IsSecretNotFoundError(errors.New("fake error")); ok {
		t.Error("IsSecretNotFoundError returned valid for invalid error")
	}
}
