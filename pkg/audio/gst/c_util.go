package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>
*/
import "C"

// init runs gst_init at luanch
func init() {
	C.gst_init(nil, nil)
}

// gobool provides an easy type conversion between C.gboolean and a go bool.
func gobool(b C.gboolean) bool {
	return b != 0
}

// gboolean converts a go bool to a C.gboolean.
func gboolean(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

// structureToGoMap converts a GstStructure into a Go map of strings.
func structureToGoMap(st *C.GstStructure) map[string]string {
	goDetails := make(map[string]string)
	numFields := int(C.gst_structure_n_fields((*C.GstStructure)(st)))
	for i := 0; i < numFields-1; i++ {
		fieldName := C.gst_structure_nth_field_name((*C.GstStructure)(st), (C.guint)(i))
		fieldValue := C.gst_structure_get_value((*C.GstStructure)(st), (*C.gchar)(fieldName))
		strValueDup := C.g_strdup_value_contents((*C.GValue)(fieldValue))
		goDetails[C.GoString(fieldName)] = C.GoString(strValueDup)
	}
	return goDetails
}
