package errors

import (
	"time"
)

type RequeueError struct {
	errMsg          string
	requeueDuration time.Duration
}

func (r *RequeueError) Error() string {
	return r.errMsg
}

func (r *RequeueError) Duration() time.Duration {
	return r.requeueDuration
}

func NewRequeueError(msg string, requeueSeconds int) error {
	return &RequeueError{
		errMsg:          msg,
		requeueDuration: time.Second * time.Duration(requeueSeconds),
	}
}

func IsRequeueError(err error) (*RequeueError, bool) {
	if qerr, ok := err.(*RequeueError); ok {
		return qerr, true
	}
	return nil, false
}
