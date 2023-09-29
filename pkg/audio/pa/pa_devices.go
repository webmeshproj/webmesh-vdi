//go:build audio

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

package pa

/*
#cgo pkg-config: libpulse
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <pulse/context.h>
#include <pulse/introspect.h>

extern void successCb(int success, void *userdata);

void device_success_cb(pa_context *c, int success, void *userdata)
{
	successCb(success, userdata);
};
*/
import "C"

import (
	"fmt"

	gopointer "github.com/mattn/go-pointer"
)

// Device is a go wrapper around a loaded pulse audio module.
// It contains internal pointers to the context that created it
// as well as additional metadata.
type device struct {
	pulseCtx          *C.pa_context
	id                int
	name, description string
	unloaded          bool
}

// nativeCtx returns the native pulse-audio context backing this instance.
func (d *device) nativeCtx() *C.pa_context { return d.pulseCtx }

// ID returns the ID of this module.
func (d *device) ID() int { return d.id }

// Name returns the name of this module.
func (d *device) Name() string { return d.name }

// Description returns the description for this module.
func (d *device) Description() string { return d.description }

// IsUnloaded returns true if this module has been unloaded.
func (d *device) IsUnloaded() bool { return d.unloaded }

// Unload will unload this pulse device.
func (d *device) Unload() error {
	resChan, chPtr := newSuccessChan()
	defer gopointer.Unref(chPtr)

	// start the operation
	op := C.pa_context_unload_module(
		(*C.pa_context)(d.nativeCtx()),
		(C.uint32_t)(d.ID()),
		C.pa_context_success_cb_t(C.device_success_cb),
		chPtr,
	)

	// Wait and return the outcome
	errFailed := fmt.Errorf("Failed to remove device '%s' with id: %d", d.Name(), d.ID())
	if op == nil {
		return errFailed
	}
	defer C.pa_operation_unref((*C.pa_operation)(op))

	// wait for the operation to finish
	if err := waitForFinish(op, resChan, errFailed); err != nil {
		return err
	}

	// set this device as unloaded
	d.unloaded = true
	return nil
}
