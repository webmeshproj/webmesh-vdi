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

// Caps is a wrapper around C.GstCaps. It provides an internal function for easy type
// conversion.
type Caps struct {
	Type string
	Data map[string]interface{}
}

func (g *Caps) toCGstCaps() (*C.GstCaps, error) {
	var structStr string
	structStr = g.Type
	if g.Data != nil {
		elems := make([]string, 0)
		for k, v := range g.Data {
			elems = append(elems, fmt.Sprintf("%s=%v", k, v))
		}
		structStr = fmt.Sprintf("%s, %s", g.Type, strings.Join(elems, ", "))
	}
	cstr := C.CString(structStr)
	defer C.free(unsafe.Pointer(cstr))
	p := C.malloc(C.size_t(128))
	defer C.free(p)
	cstruct := C.gst_structure_from_string((*C.gchar)(cstr), (**C.gchar)(p))
	if cstruct == nil {
		return nil, errors.New("Could not create GstStructure from Structure")
	}

	caps := C.gst_caps_new_empty()
	if caps == nil {
		return nil, errors.New("Could not create new empty caps")
	}
	C.gst_caps_append_structure((*C.GstCaps)(caps), (*C.GstStructure)(cstruct))

	return caps, nil
}
