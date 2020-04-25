package errors

import (
	"errors"
	"testing"
	"time"
)

func TestRequeueError(t *testing.T) {
	qerr := NewRequeueError("fake reason", 3)
	if qerr.Error() != "fake reason" {
		t.Error("Error body is malformed")
	}
	if qerr.Duration() != time.Duration(3)*time.Second {
		t.Error("Duration is malformed")
	}

	if aerr, ok := IsRequeueError(qerr); !ok {
		t.Error("Should be a valid requeue error")
	} else if aerr != qerr {
		t.Error("Returned error isn't pointer to the same error")
	}

	if _, ok := IsRequeueError(errors.New("fake error")); ok {
		t.Error("IsRequeuError returned valid for invalid error")
	}
}
