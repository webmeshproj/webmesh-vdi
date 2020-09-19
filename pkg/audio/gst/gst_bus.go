package gst

// #cgo pkg-config: gstreamer-1.0
// #cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
// #include <gst/gst.h>
import "C"
import (
	"errors"
	"unsafe"
)

// Bus is a Go wrapper around a GstBus. It provides convenience methods for
// unref-ing and popping messages from the queue.
type Bus struct {
	bus *C.GstBus
}

// BusFromPipeline returns a new Bus object for the GstBus on the given GstPipeline.
func BusFromPipeline(pipeline *C.GstElement) (*Bus, error) {
	bus := C.gst_element_get_bus((*C.GstElement)(pipeline))
	if bus == nil {
		return nil, errors.New("Could not retrieve bus from pipeline")
	}
	return &Bus{bus: bus}, nil
}

// Native returns the native pointer to the GstBus.
func (b *Bus) Native() unsafe.Pointer { return unsafe.Pointer(b.bus) }

// BlockPopMessage blocks until a message is available on the bus and then returns in.
func (b *Bus) BlockPopMessage() *Message {
	msg := C.gst_bus_timed_pop_filtered(
		(*C.GstBus)(b.bus),
		C.GST_CLOCK_TIME_NONE,
		C.GST_MESSAGE_ANY,
	)
	if msg == nil {
		return nil
	}
	return NewMessage(msg)
}

// Unref wraps `gst_object_unref` on the underlying GstBus.
func (b *Bus) Unref() {
	C.gst_object_unref((C.gpointer)(b.bus))
}
