package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>

void cgo_g_object_set_boolean (GObject * obj, gchar * fieldName, gboolean value) {
		  g_object_set (obj, fieldName, value, NULL);
}

void cgo_g_object_set_string (GObject * obj, gchar * fieldName, gchar * value) {
	  g_object_set (obj, fieldName, value, NULL);
}

void cgo_g_object_set_int (GObject * obj, gchar * fieldName, gint value) {
	  g_object_set (obj, fieldName, value, NULL);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

// Element is a Go wrapper around a GstElement. This is intended to be used with plugins
// and not pipelines.
type Element struct {
	elem *C.GstElement
}

// NewElement is a generic wrapper around `gst_element_factory_make`.
func NewElement(name string) (*Element, error) {
	elemName := C.CString(name)
	defer C.free(unsafe.Pointer(elemName))
	elem := C.gst_element_factory_make((*C.gchar)(elemName), nil)
	if elem == nil {
		return nil, fmt.Errorf("Could not create element: %s", name)
	}
	return &Element{elem: elem}, nil
}

// NewElementMany is a convenience wrapper around building many GstElements in a
// single function call. It returns an error if the creation of any element fails. A
// map is returned with keys matching the names provided as arguments.
func NewElementMany(elemNames ...string) (map[string]*Element, error) {
	elemMap := make(map[string]*Element)
	for _, name := range elemNames {
		elem, err := NewElement(name)
		if err != nil {
			return nil, err
		}
		elemMap[name] = elem
	}
	return elemMap, nil
}

// native returns the underlying GstElement.
func (e *Element) native() *C.GstElement { return e.elem }

// Name returns the name of this GstElement.
func (e *Element) Name() string { return C.GoString((*C.GstObject)(unsafe.Pointer(e.elem)).name) }

// Set sets fieldName to fieldValue on the underlying GstElement.
func (e *Element) Set(fieldName string, fieldValue interface{}) error {
	cfieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cfieldName))
	switch reflect.TypeOf(fieldValue).Kind() {
	case reflect.String:
		cval := C.CString(fieldValue.(string))
		defer C.free(unsafe.Pointer(cval))
		C.cgo_g_object_set_string(
			(*C.GObject)(unsafe.Pointer(e.elem)),
			(*C.gchar)(cfieldName),
			(*C.gchar)(cval),
		)
	case reflect.Int:
		cval := C.gint(fieldValue.(int))
		C.cgo_g_object_set_int(
			(*C.GObject)(unsafe.Pointer(e.elem)),
			(*C.gchar)(cfieldName),
			(C.gint)(cval),
		)

	case reflect.Bool:
		cval := gboolean(fieldValue.(bool))
		C.cgo_g_object_set_boolean(
			(*C.GObject)(unsafe.Pointer(e.elem)),
			(*C.gchar)(cfieldName),
			(C.gboolean)(cval),
		)

	default:
		return fmt.Errorf("Unhandled type for Element.Set(): %s", reflect.TypeOf(fieldValue).String())
	}
	return nil
}

func newEmptyPipeline() (*C.GstElement, error) {
	pipelineName := C.CString(time.Now().String())
	defer C.free(unsafe.Pointer(pipelineName))

	pipeline := C.gst_pipeline_new((*C.gchar)(pipelineName))
	if pipeline == nil {
		return nil, errors.New("Could not create new pipeline")
	}
	return pipeline, nil
}

func newPipelineFromString(launchv string) (*C.GstElement, error) {
	cLaunchv := C.CString(launchv)
	defer C.free(unsafe.Pointer(cLaunchv))
	var gerr *C.GError
	pipeline := C.gst_parse_launch((*C.gchar)(cLaunchv), (**C.GError)(&gerr))
	if gerr != nil {
		defer C.g_error_free((*C.GError)(gerr))
		errMsg := C.GoString(gerr.message)
		return nil, errors.New(errMsg)
	}
	return pipeline, nil
}
