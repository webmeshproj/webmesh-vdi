package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>
#include <gst/app/gstappsink.h>
*/
import "C"

func init() {
	C.gst_init(nil, nil)
}

func gobool(b C.gboolean) bool {
	return b != 0
}
