package errors

import "testing"

func TestNew(t *testing.T) {
	var err error
	if err = New("error"); err == nil {
		t.Error("New should return a new error")
	}
}
