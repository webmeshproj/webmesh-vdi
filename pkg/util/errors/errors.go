package errors

import (
	goerrors "errors"
)

// New wraps the stdlib errors.New for simplicity when using this package.
func New(msg string) error {
	return goerrors.New(msg)
}
