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
#include <stdlib.h>
#include <pulse/context.h>
#include <pulse/introspect.h>
#include <pulse/thread-mainloop.h>

extern void success_cb (pa_context *c, int success, void *userdata);
extern void state_change_cb (pa_context *c, void *userdata);
extern void new_module_cb (pa_context *c, uint32_t idx, void *userdata);
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

// DeviceManagerOpts represent options to pass to the device manager.
type DeviceManagerOpts struct {
	PulseServer string
}

// DeviceManager is an object for managing virtual PulseAudio devices.
type DeviceManager struct {
	server   string
	devices  []*Device
	state    C.pa_context_state_t
	mainLoop *C.pa_threaded_mainloop
	paCtx    *C.pa_context
	selfPtr  unsafe.Pointer
	mux      sync.Mutex
}

// NewDeviceManager returns a new DeviceManager.
func NewDeviceManager(opts *DeviceManagerOpts) (*DeviceManager, error) {
	devManager := &DeviceManager{
		server:   opts.PulseServer,
		devices:  make([]*Device, 0),
		mainLoop: C.pa_threaded_mainloop_new(),
	}
	if err := devManager.connect(); err != nil {
		devManager.disconnect()
		return nil, err
	}
	return devManager, nil
}

// connect will create a new context and start the main loop.
func (p *DeviceManager) connect() error {
	// build args
	cname := C.CString(fmt.Sprintf("device_manager_%s", time.Now().String()))
	defer C.free(unsafe.Pointer(cname))
	cserverName := C.CString(p.getServer())
	defer C.free(unsafe.Pointer(cserverName))

	// build a new context
	p.paCtx = C.pa_context_new((*C.pa_mainloop_api)(p.getMainLoopAPI()), (*C.char)(cname))
	p.selfPtr = gopointer.Save(p)

	// set the state callback on the context
	C.pa_context_set_state_callback(
		(*C.pa_context)(p.nativeCtx()),
		C.pa_context_notify_cb_t(C.state_change_cb),
		p.selfPtr,
	)

	// Start the connection. This call is asynchronous.
	ret := C.pa_context_connect(
		(*C.pa_context)(p.nativeCtx()),
		(*C.char)(cserverName),
		C.PA_CONTEXT_NOFLAGS, nil,
	)
	if ret < 0 {
		return errors.New("Could not start connection to pulse server")
	}

	// start the main loop
	if ret := C.pa_threaded_mainloop_start((*C.pa_threaded_mainloop)(p.getMainLoop())); ret != 0 {
		return errors.New("Could not start threaded main loop for pa server")
	}

	return nil
}

// getState returns the `pa_context_state_t` of the current context.
func (p *DeviceManager) getState() C.pa_context_state_t { return p.state }

// stateChanged is fired everytime there is a change to the underlying pulse context.
// It sets the current state locally, and a switch statement is templated out for further
// fine-grained control.
func (p *DeviceManager) stateChanged() {
	p.state = C.pa_context_get_state((*C.pa_context)(p.nativeCtx()))
	switch p.state {
	case C.PA_CONTEXT_UNCONNECTED:
	case C.PA_CONTEXT_CONNECTING:
	case C.PA_CONTEXT_READY:
	case C.PA_CONTEXT_FAILED:
	case C.PA_CONTEXT_TERMINATED:
	case C.PA_CONTEXT_AUTHORIZING:
	case C.PA_CONTEXT_SETTING_NAME:
	}
}

// disconnect will disconnect from the pulse audio server and free
// all associated resources.
func (p *DeviceManager) disconnect() {
	if p.getState() == C.PA_CONTEXT_READY {
		C.pa_context_disconnect((*C.pa_context)(p.nativeCtx()))
	}
	C.pa_context_unref((*C.pa_context)(p.nativeCtx()))
	gopointer.Unref(p.selfPtr)
	C.pa_threaded_mainloop_stop((*C.pa_threaded_mainloop)(p.getMainLoop()))
	C.pa_threaded_mainloop_free((*C.pa_threaded_mainloop)(p.getMainLoop()))
}

// nativeCtx returns the native `pa_context`.
func (p *DeviceManager) nativeCtx() *C.pa_context { return p.paCtx }

// getMainLoop returns the mainloop for this context.
func (p *DeviceManager) getMainLoop() *C.pa_threaded_mainloop { return p.mainLoop }

// getMainLoopAPI returns the `pa_mainloop_api` object for this context's mainloop.
func (p *DeviceManager) getMainLoopAPI() *C.pa_mainloop_api {
	return C.pa_threaded_mainloop_get_api(p.getMainLoop())
}

// getServer returns the path to the current pulse audio server.
func (p *DeviceManager) getServer() string { return p.server }

// appendDevice adds the given device to the internal memory of devices.
func (p *DeviceManager) appendDevice(device *Device) { p.devices = append(p.devices, device) }

// loadMoudle is a synchronous go wrapper around `pa_context_load_module`.
func (p *DeviceManager) loadModule(name, args string) (int, error) {
	// setup arguments
	cModType := C.CString(name)
	cModArgs := C.CString(args)
	defer C.free(unsafe.Pointer(cModType))
	defer C.free(unsafe.Pointer(cModArgs))

	resChan, chPtr := newIndexChan()
	defer gopointer.Unref(chPtr)

	// start the operation
	op := C.pa_context_load_module(
		(*C.pa_context)(p.nativeCtx()),
		(*C.char)(cModType),
		(*C.char)(cModArgs),
		C.pa_context_index_cb_t(C.new_module_cb),
		unsafe.Pointer(chPtr),
	)

	failErr := fmt.Errorf("Failed to load module %s with args %s", name, args)
	if op == nil {
		return 0, failErr
	}

	defer C.pa_operation_unref((*C.pa_operation)(op))
	return waitForIndexFinish(op, resChan, failErr)
}

// AddSink adds a new null-sink with the given name and description.
func (p *DeviceManager) AddSink(name, description string) (*Device, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	args := fmt.Sprintf(`sink_name="%s" sink_properties=device.description="%s"`, name, description)
	deviceID, err := p.loadModule("module-null-sink", args)
	if err != nil {
		return nil, err
	}
	device := &Device{
		pulseCtx:    p.nativeCtx(),
		id:          deviceID,
		name:        name,
		description: description,
	}
	p.appendDevice(device)
	return device, nil
}

// SourceOpts represents options for a creating a new virtual source.
type SourceOpts struct {
	Name                 string
	Description          string
	FifoPath             string
	SampleFormat         string
	Channels, SampleRate int
}

// AddSource adds a new pipe-source with the given name, description, and FIFO path.
func (p *DeviceManager) AddSource(opts *SourceOpts) (*Device, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	args := fmt.Sprintf(`source_name="%s" source_properties=device.description="%s" file="%s" format="%s" rate=%d channels=%d`,
		opts.Name, opts.Description, opts.FifoPath, opts.SampleFormat, opts.SampleRate, opts.Channels,
	)
	deviceID, err := p.loadModule("module-pipe-source", args)
	if err != nil {
		return nil, err
	}
	device := &Device{
		pulseCtx:    p.nativeCtx(),
		id:          deviceID,
		name:        opts.Name,
		description: opts.Description,
	}
	p.appendDevice(device)
	return device, nil
}

// SetDefaultSource will set the default source for recording clients.
func (p *DeviceManager) SetDefaultSource(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	// create a channel for the response
	resChan := make(chan bool)

	// save a pointer to the channel
	chPtr := gopointer.Save(resChan)
	defer gopointer.Unref(chPtr)

	op := C.pa_context_set_default_source(
		(*C.pa_context)(p.nativeCtx()),
		(*C.char)(cName),
		C.pa_context_success_cb_t(C.success_cb), chPtr,
	)

	errFailed := fmt.Errorf("Failed to set %s as the default source", name)

	if op == nil {
		return errFailed
	}

	defer C.pa_operation_unref((*C.pa_operation)(op))

	return waitForFinish(op, resChan, errFailed)
}

// Devices returns a list of the current devices managed by this instance.
func (p *DeviceManager) Devices() []*Device { return p.devices }

// WaitForReady waits for the DeviceManager to be able to execute operations
// against the PulseAudio server. Since all calls are async, this method SHOULD
// be run after a new DeviceManager is created.
func (p *DeviceManager) WaitForReady(timeout time.Duration) error {
	var ctx context.Context
	var cancel func()
	if timeout > 1 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return errors.New("Context deadline exceeded")
		default:
			switch p.getState() {
			case C.PA_CONTEXT_READY:
				return nil
			case C.PA_CONTEXT_FAILED:
				return errors.New("PA context failed to connect")
			case C.PA_CONTEXT_TERMINATED:
				return errors.New("PA context was terminated")
			}
		}
	}
}

// Destroy will unload all currently managed PA devices and close the connection
// to the pulse server.
func (p *DeviceManager) Destroy() error {
	p.mux.Lock()
	defer p.mux.Unlock()
	for _, dev := range p.Devices() {
		if err := dev.Unload(); err != nil {
			return err
		}
	}
	p.disconnect()
	return nil
}
