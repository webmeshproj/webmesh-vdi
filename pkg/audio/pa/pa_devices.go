package pa

/*
#cgo pkg-config: libpulse
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <pulse/context.h>
#include <pulse/introspect.h>

extern void success_cb (pa_context *c, int success, void *userdata);
*/
import "C"

import (
	"fmt"

	gopointer "github.com/mattn/go-pointer"
)

// Device is a go wrapper around a loaded pulse audio module.
// It contains internal pointers to the context that created it
// as well as additional metadata.
type Device struct {
	pulseCtx          *C.pa_context
	id                int
	name, description string
	unloaded          bool
}

// nativeCtx returns the native pulse-audio context backing this instance.
func (d *Device) nativeCtx() *C.pa_context { return d.pulseCtx }

// ID returns the ID of this module.
func (d *Device) ID() int { return d.id }

// Name returns the name of this module.
func (d *Device) Name() string { return d.name }

// Description returns the description for this module.
func (d *Device) Description() string { return d.description }

// IsUnloaded returns true if this module has been unloaded.
func (d *Device) IsUnloaded() bool { return d.unloaded }

// Unload will unload this pulse device.
func (d *Device) Unload() error {
	resChan, chPtr := newSuccessChan()
	defer gopointer.Unref(chPtr)

	// start the operation
	op := C.pa_context_unload_module(
		(*C.pa_context)(d.nativeCtx()),
		(C.uint32_t)(d.ID()),
		C.pa_context_success_cb_t(C.success_cb),
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
