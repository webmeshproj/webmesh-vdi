/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package errors

import (
	"encoding/json"
	"io"
	"net/http"
)

// ErrorStatus represents a type of API error
type ErrorStatus string

// Error types
const (
	Unauthorized ErrorStatus = "Unauthorized"
	Forbidden    ErrorStatus = "Forbidden"
	NotFound     ErrorStatus = "NotFound"
	ServerError  ErrorStatus = "ServerError"
)

// APIError is for errors from the API server. It's main purpose
// is to provide a quick interface for returning json encoded error
// messages
type APIError struct {
	// A message describing the error
	ErrMsg string `json:"error"`
	// The status for the error.
	ErrStatus ErrorStatus `json:"status"`
}

// CheckAPIError evaluates if the HTTP response contains an API error.
// If so, an attempt is made to unmarshal it into an API error. If this
// fails, then an error containing the original body is returned, or any
// error from attempting to read the body.
func CheckAPIError(r *http.Response) error {
	if r.StatusCode == http.StatusOK {
		return nil
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var out APIError
	if err := json.Unmarshal(body, &out); err != nil {
		return New(string(body))
	}
	return &out
}

// Error implements the error interface
func (r *APIError) Error() string {
	return r.ErrMsg
}

// ToAPIError converts a generic error into an API error
func ToAPIError(err error, errStatus ErrorStatus) *APIError {
	return &APIError{
		ErrMsg:    err.Error(),
		ErrStatus: errStatus,
	}
}

// JSON returns the json encoded error. Error checking is skipped since
// this is only used internally and for valid strings.
func (r *APIError) JSON() []byte {
	out, _ := json.MarshalIndent(r, "", "    ")
	return out
}

// IsAPINotFound checks if the given error from the API is a NotFound error.
func IsAPINotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.ErrStatus == NotFound {
			return true
		}
	}
	return false
}

// IsAPIUnauthorized checks if the given error from the API is a Unauthorized error.
func IsAPIUnauthorized(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.ErrStatus == Unauthorized {
			return true
		}
	}
	return false
}

// IsAPIForbidden checks if the given error from the API is a Forbidden error.
func IsAPIForbidden(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.ErrStatus == Forbidden {
			return true
		}
	}
	return false
}

// IsAPIServerError checks if the given error from the API is a ServerError error.
func IsAPIServerError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.ErrStatus == ServerError {
			return true
		}
	}
	return false
}
