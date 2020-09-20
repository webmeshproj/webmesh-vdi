package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>
*/
import "C"

// Bus is a Go wrapper around a GstBus. It provides convenience methods for
// unref-ing and popping messages from the queue.
type Bus struct {
	pipeline *Pipeline
	bus      *C.GstBus
}

// native returns the underlying GstBus.
func (b *Bus) native() *C.GstBus { return b.bus }

// BlockPopMessage blocks until a message is available on the bus and then returns it.
// If the underlying pipeline is stopped while this function is being called, it will
// return nil.
func (b *Bus) BlockPopMessage() *Message {
	for {
		msg := C.gst_bus_timed_pop_filtered(
			(*C.GstBus)(b.native()),
			C.GST_SECOND,
			C.GST_MESSAGE_ANY,
		)
		if msg == nil {
			if b.pipeline.IsClosed() {
				break
			}
			continue
		}
		return NewMessage(msg)
	}
	return nil
}

// Unref wraps `gst_object_unref` on the underlying GstBus.
func (b *Bus) Unref() {
	C.gst_object_unref((C.gpointer)(b.native()))
}
