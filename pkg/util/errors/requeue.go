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
