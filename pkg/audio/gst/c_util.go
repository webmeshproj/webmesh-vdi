package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>
*/
import "C"
import "unsafe"

// init runs gst_init at luanch
func init() {
	C.gst_init(nil, nil)
}

// gobool provides an easy type conversion between C.gboolean and a go bool.
func gobool(b C.gboolean) bool {
	return b != 0
}

// elementGoName retrieves the name of the given element as a Go string
func elementGoName(elem *C.GstElement) string {
	return C.GoString((*C.GstObject)(unsafe.Pointer(elem)).name)
}
