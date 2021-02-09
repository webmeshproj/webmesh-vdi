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
