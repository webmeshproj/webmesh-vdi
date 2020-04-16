package errors

import "encoding/json"

type APIError struct {
	ErrMsg string `json:"error"`
}

func (r *APIError) Error() string {
	return r.ErrMsg
}

func ToAPIError(err error) *APIError {
	return &APIError{
		ErrMsg: err.Error(),
	}
}

func (r *APIError) JSON() []byte {
	out, _ := json.MarshalIndent(r, "", "    ")
	return out
}
