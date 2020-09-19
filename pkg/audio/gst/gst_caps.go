package gst

// #cgo pkg-config: gstreamer-1.0
// #cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
// #include <gst/gst.h>
import "C"

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

// Caps is a wrapper around GstCaps. It provides a function for easy type
// conversion.
type Caps struct {
	Type string
	Data map[string]interface{}
}

// NewRawCaps returns new GstCaps with the given format, sample-rate, and channels.
func NewRawCaps(format string, rate, channels int) *Caps {
	caps := &Caps{
		Type: "audio/x-raw",
		Data: map[string]interface{}{
			"format":   format,
			"rate":     rate,
			"channels": channels,
		},
	}
	return caps
}

// ToGstCaps returns the GstCaps representation of this Caps instance.
func (g *Caps) ToGstCaps() (*C.GstCaps, error) {
	var structStr string
	structStr = g.Type
	// build a structure string from the data
	if g.Data != nil {
		elems := make([]string, 0)
		for k, v := range g.Data {
			elems = append(elems, fmt.Sprintf("%s=%v", k, v))
		}
		structStr = fmt.Sprintf("%s, %s", g.Type, strings.Join(elems, ", "))
	}
	// convert the structure string to a cstring
	cstr := C.CString(structStr)
	defer C.free(unsafe.Pointer(cstr))
	// a small buffer for garbage
	p := C.malloc(C.size_t(128))
	defer C.free(p)
	// create a structure from the string
	cstruct := C.gst_structure_from_string((*C.gchar)(cstr), (**C.gchar)(p))
	if cstruct == nil {
		return nil, errors.New("Could not create GstStructure from Structure")
	}
	// create a new empty caps object
	caps := C.gst_caps_new_empty()
	if caps == nil {
		return nil, errors.New("Could not create new empty caps")
	}

	// append the structure to the caps
	C.gst_caps_append_structure((*C.GstCaps)(caps), (*C.GstStructure)(cstruct))

	return caps, nil
}
