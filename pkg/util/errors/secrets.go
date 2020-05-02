package errors

import "fmt"

// The error message format for a SecretNotFoundError
const secretNotFoundFormat = "Secret '%s' could not be found"

// SecretNotFoundError is used to signal from a secrets backend that the requested
// secret does not exist.
type SecretNotFoundError struct {
	errMsg string
}

// Error implements the error interface
func (r *SecretNotFoundError) Error() string {
	return r.errMsg
}

// NewSecretNotFoundError returns a new SecretNotFoundError for the given resource
// name.
func NewSecretNotFoundError(secret string) error {
	return &SecretNotFoundError{
		errMsg: fmt.Sprintf(secretNotFoundFormat, secret),
	}
}

// IsSecretNotFoundError returns true if the given error is a SecretNotFoundError.
func IsSecretNotFoundError(err error) bool {
	if _, ok := err.(*SecretNotFoundError); ok {
		return true
	}
	return false
}
