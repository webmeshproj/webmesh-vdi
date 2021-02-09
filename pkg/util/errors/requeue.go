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
	"time"
)

// RequeueError is used to float errors back up to the controller signalling
// that an error has not actually ocurred, but instead the reconcile attempt
// should be requeued due to a resource not being ready.
type RequeueError struct {
	// the message that will be logged by the controller
	errMsg string
	// the amount of time to wait before the reconcile should be requeued
	requeueDuration time.Duration
}

// Error implements the error interface
func (r *RequeueError) Error() string {
	return r.errMsg
}

// Duration returns the wait time for this error
func (r *RequeueError) Duration() time.Duration {
	return r.requeueDuration
}

// NewRequeueError returns a new requeue error. The provided string will be logged,
// and the reconcile will be requeued after the given number of seconds.
func NewRequeueError(msg string, requeueSeconds int) *RequeueError {
	return &RequeueError{
		errMsg:          msg,
		requeueDuration: time.Second * time.Duration(requeueSeconds),
	}
}

// IsRequeueError returns true if the provided error interface is a RequeueError.
// If it is a requeue error, return the underlying object.
func IsRequeueError(err error) (*RequeueError, bool) {
	if qerr, ok := err.(*RequeueError); ok {
		return qerr, true
	}
	return nil, false
}
