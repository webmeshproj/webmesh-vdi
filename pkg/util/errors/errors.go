package errors

import (
	goerrors "errors"
)

func New(msg string) error {
	return goerrors.New(msg)
}
