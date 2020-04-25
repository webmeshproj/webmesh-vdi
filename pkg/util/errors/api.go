package errors

import "encoding/json"

// APIError is for errors from the API server. It's main purpose
// is to provide a quick interface for returning json encoded error
// messages
type APIError struct {
	// A message describing the error
	ErrMsg string `json:"error"`
}

// Error implements the error interface
func (r *APIError) Error() string {
	return r.ErrMsg
}

// ToAPIError converts a generic error into an API error
func ToAPIError(err error) *APIError {
	return &APIError{
		ErrMsg: err.Error(),
	}
}

// JSON returns the json encoded error. Error checking is skipped since
// this is only used internally and for valid strings.
func (r *APIError) JSON() []byte {
	out, _ := json.MarshalIndent(r, "", "    ")
	return out
}
