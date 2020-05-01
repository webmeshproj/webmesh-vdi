package errors

import "fmt"

const secretNotFoundFormat = "Secret '%s' could not be found"

type SecretNotFoundError struct {
	errMsg string
}

func (r *SecretNotFoundError) Error() string {
	return r.errMsg
}

func NewSecretNotFoundError(secret string) error {
	return &SecretNotFoundError{
		errMsg: fmt.Sprintf(secretNotFoundFormat, secret),
	}
}

func IsSecretNotFoundError(err error) bool {
	if _, ok := err.(*SecretNotFoundError); ok {
		return true
	}
	return false
}
