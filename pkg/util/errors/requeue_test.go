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
	"errors"
	"testing"
	"time"
)

func TestRequeueError(t *testing.T) {
	qerr := NewRequeueError("fake reason", 3)
	if qerr.Error() != "fake reason" {
		t.Error("Error body is malformed")
	}
	if qerr.Duration() != time.Duration(3)*time.Second {
		t.Error("Duration is malformed")
	}

	if aerr, ok := IsRequeueError(qerr); !ok {
		t.Error("Should be a valid requeue error")
	} else if aerr != qerr {
		t.Error("Returned error isn't pointer to the same error")
	}

	if _, ok := IsRequeueError(errors.New("fake error")); ok {
		t.Error("IsRequeuError returned valid for invalid error")
	}
}
