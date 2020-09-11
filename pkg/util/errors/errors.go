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
