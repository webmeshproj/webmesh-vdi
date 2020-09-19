package gst

// #cgo pkg-config: gstreamer-1.0
// #cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
// #include <gst/gst.h>
import "C"

import (
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func newPipeline() (*C.GstElement, error) {
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

func newFifoSink(fifoName string) (*C.GstElement, error) {
	fifoSink, err := newElement("filesink")
	if err != nil {
		return nil, err
	}
	gsink := glib.Take(unsafe.Pointer(fifoSink))
	if err := gsink.Set("location", fifoName); err != nil {
		return nil, err
	}
	if err := gsink.Set("append", true); err != nil {
		return nil, err
	}
	return fifoSink, nil
}

func newFdSink(fd int) (*C.GstElement, error) {
	fdSink, err := newElement("fdsink")
	if err != nil {
		return nil, err
	}
	gsink := glib.Take(unsafe.Pointer(fdSink))
	if err := gsink.Set("fd", fd); err != nil {
		return nil, err
	}
	return fdSink, nil
}

func newFdSrc(fd int) (*C.GstElement, error) {
	fdSrc, err := newElement("fdsrc")
	if err != nil {
		return nil, err
	}
	gsink := glib.Take(unsafe.Pointer(fdSrc))
	if err := gsink.Set("fd", fd); err != nil {
		return nil, err
	}
	return fdSrc, nil
}

func newPulseSrc(opts *PlaybackPipelineOpts) (*C.GstElement, error) {
	pulseSrc, err := newElement("pulsesrc")
	if err != nil {
		return nil, err
	}

	gsrc := glib.Take(unsafe.Pointer(pulseSrc))
	if err := gsrc.Set("server", opts.PulseServer); err != nil {
		return nil, err
	}
	if err := gsrc.Set("device", opts.DeviceName); err != nil {
		return nil, err
	}

	return pulseSrc, nil
}

func newPulseCaps(format string, rate, channels int) (*C.GstCaps, error) {
	caps := &Caps{
		Type: "audio/x-raw",
		Data: map[string]interface{}{
			"format":   format,
			"rate":     rate,
			"channels": channels,
		},
	}
	return caps.toCGstCaps()
}

func newElement(name string) (*C.GstElement, error) {
	elemName := C.CString(name)
	defer C.free(unsafe.Pointer(elemName))
	elem := C.gst_element_factory_make((*C.gchar)(elemName), nil)
	if elem == nil {
		return nil, fmt.Errorf("Could not create element: %s", name)
	}
	return elem, nil
}
