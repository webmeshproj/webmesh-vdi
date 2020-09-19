package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>
*/
import "C"

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/go-logr/logr"
)

// Pipeline is the base implementation of a GstPipeline using CGO to wrap
// gstreamer API calls. It provides methods to be inherited by the extending
// PlaybackPipeline and RecordingPipeline objects. The struct itself implements
// a ReadWriteCloser.
type Pipeline struct {
	// A logger interface for messages
	logger logr.Logger
	// The underlying pointer to the C pipeline element
	pipelineElement *C.GstElement
	// The piped (unbuffered) reader
	reader *os.File
	// The piped (unbuffered) writer
	writer *os.File
	// A buffer wrapping the reader. For use in the Read method.
	rbuf io.Reader
	// A buffer wrapping the writer. For use in the Write method.
	wbuf io.Writer
	// A channel for signaling the message poller to stop
	stopCh chan struct{}
	// errors contains any errors reported back from the pipeline
	errors []error
	// the current state of the pipeline
	currentState string
}

// NewPipeline builds and returns a new CPipeline instance.
func NewPipeline(logger logr.Logger) (*Pipeline, error) {
	pipeline, err := newPipeline()
	if err != nil {
		return nil, err
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &Pipeline{
		logger:          logger,
		pipelineElement: pipeline,
		reader:          r,
		writer:          w,
		rbuf:            bufio.NewReader(r),
		wbuf:            bufio.NewWriter(w),
		stopCh:          make(chan struct{}),
		errors:          make([]error, 0),
	}, nil
}

// NewPipelineFromLaunchString returns a new GstPipeline from the given launch string. If useFdSrc or useFdSink
// are true, then the pipeline string is additionally formatted with the internal read/write buffers.
func NewPipelineFromLaunchString(logger logr.Logger, launchStr string, useFdSrc, useFdSink bool) (*Pipeline, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	if useFdSrc {
		launchStr = fmt.Sprintf("fdsrc fd=%d ! %s", int(r.Fd()), launchStr)
	}
	if useFdSink {
		launchStr = fmt.Sprintf("%s ! fdsink fd=%d", launchStr, int(w.Fd()))
	}
	pipeline, err := newPipelineFromString(launchStr)
	if err != nil {
		return nil, err
	}
	return &Pipeline{
		logger:          logger,
		pipelineElement: pipeline,
		reader:          r,
		writer:          w,
		rbuf:            bufio.NewReader(r),
		wbuf:            bufio.NewWriter(w),
		stopCh:          make(chan struct{}),
		errors:          make([]error, 0),
	}, nil
}

// Native returns the pointer to the underlying pipeline element.
func (p *Pipeline) Native() *C.GstElement { return p.pipelineElement }

// Read implements a Reader and returns data from the read buffer.
func (p *Pipeline) Read(b []byte) (int, error) { return p.rbuf.Read(b) }

// ReaderFd returns the file descriptor for the read buffer.
func (p *Pipeline) ReaderFd() uintptr { return p.reader.Fd() }

// Write implements a Writer and places data in the write buffer.
func (p *Pipeline) Write(b []byte) (int, error) { return p.wbuf.Write(b) }

// WriterFd returns the file descriptor for the write buffer.
func (p *Pipeline) WriterFd() uintptr { return p.writer.Fd() }

// Start will start the underlying pipeline.
func (p *Pipeline) Start() error {
	p.logger.Info("Starting pipeline")
	if err := p.setupPipelineBus(); err != nil {
		return err
	}
	stateRet := C.gst_element_set_state((*C.GstElement)(p.pipelineElement), C.GST_STATE_PLAYING)
	if stateRet == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("Failed to start pipeline")
	}
	return nil
}

// IsRunning will return true if the gstreamer pipeline is currently running.
func (p *Pipeline) IsRunning() bool {
	return p.pipelineElement != nil && p.currentState != "NULL"
}

// Close implements a Closer and closes the read and write pipes.
func (p *Pipeline) Close() error {
	stateRet := C.gst_element_set_state((*C.GstElement)(p.pipelineElement), C.GST_STATE_NULL)
	if stateRet == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("Failed to stop pipeline")
	}
	if err := p.reader.Close(); err != nil {
		return err
	}
	if err := p.writer.Close(); err != nil {
		return err
	}
	p.stopCh <- struct{}{}
	C.free(unsafe.Pointer(p.pipelineElement))
	p.pipelineElement = nil
	return nil
}

// Errors returns any errors that happened during the pipeline.
func (p *Pipeline) Errors() []error {
	return p.errors
}

// NewElementMany is a convenience wrapper around building many *C.GstElement's in a
// single function call. It returns an error if the creation of any element fails. A
// map is returned with keys matching the names provided as arguments.
func (p *Pipeline) NewElementMany(elemNames ...string) (map[string]*C.GstElement, error) {
	elemMap := make(map[string]*C.GstElement)
	for _, name := range elemNames {
		elem, err := NewElement(name)
		if err != nil {
			return nil, err
		}
		elemMap[name] = elem
	}
	return elemMap, nil
}

// BinAddMany is a go implementation of gst_bin_add_many to compensate for the inability
// to use variadic functions in cgo.
func (p *Pipeline) BinAddMany(elems ...*C.GstElement) error {
	for _, elem := range elems {
		if ok := C.gst_bin_add((*C.GstBin)(unsafe.Pointer(p.pipelineElement)), (*C.GstElement)(elem)); !gobool(ok) {
			return fmt.Errorf("Failed to add element to pipeline: %+v", *elem)
		}
	}
	return nil
}

// ElementLinkMany is a go implementation of gst_element_link_many to compensate for
// no variadic functions in cgo.
func (p *Pipeline) ElementLinkMany(elems ...*C.GstElement) error {
	for idx, elem := range elems {
		if idx == 0 {
			// skip the first one as the loop always links previous to current
			continue
		}
		if ok := C.gst_element_link((*C.GstElement)(elems[idx-1]), (*C.GstElement)(elem)); !gobool(ok) {
			beforeName := C.GoString((*C.GstObject)(unsafe.Pointer(elems[idx-1])).name)
			afterName := C.GoString((*C.GstObject)(unsafe.Pointer(elem)).name)
			return fmt.Errorf("Failed to link %s to %s", beforeName, afterName)
		}
	}
	return nil
}

// ElementLinkFiltered is a convenience wrapper around  gst_element_link_filtered for linking caps
// between two elements.
func (p *Pipeline) ElementLinkFiltered(beforeElem, afterElem *C.GstElement, caps *C.GstCaps) error {
	if ok := C.gst_element_link_filtered((*C.GstElement)(beforeElem), (*C.GstElement)(afterElem), (*C.GstCaps)(caps)); !gobool(ok) {
		beforeName := C.GoString((*C.GstObject)(unsafe.Pointer(beforeElem)).name)
		afterName := C.GoString((*C.GstObject)(unsafe.Pointer(afterElem)).name)
		return fmt.Errorf("Failed to link %s to %s with provider caps", beforeName, afterName)
	}
	return nil
}

// setupPipelineBus spawns a goroutine that pops messages on the bus and passes them
// to the message handler.
func (p *Pipeline) setupPipelineBus() error {
	go func() {
		for {
			select {
			case <-p.stopCh:
				return
			default:
				if p.pipelineElement == nil {
					p.logger.Info("Pipeline element has been stopped")
					return
				}
				bus := C.gst_element_get_bus((*C.GstElement)(p.pipelineElement))
				if bus == nil {
					p.logger.Error(errors.New("Could not retrieve bus from pipeline"), "Stopping message queue")
					return
				}
				msg := C.gst_bus_timed_pop_filtered((*C.GstBus)(bus), C.GST_CLOCK_TIME_NONE, C.GST_MESSAGE_ANY)
				if msg != nil {
					p.handleMessage(msg)
				}
				C.gst_object_unref((C.gpointer)(bus))
			}
		}
	}()
	return nil
}

// handleMessage handles a GstMessage on the pipeline bus.
func (p *Pipeline) handleMessage(msg *C.GstMessage) {
	// free up the message after processing
	defer C.gst_message_unref((*C.GstMessage)(msg))

	switch msg._type {

	case C.GST_MESSAGE_STREAM_START:
		p.logger.Info("Stream has started, audio data is available on the buffer")

	case C.GST_MESSAGE_EOS:
		p.logger.Info("Stream has ended, closing pipeline")
		p.Close()

	// Fires rarely
	// TODO: Parse these messages
	case C.GST_MESSAGE_INFO:
		p.logger.Info("Got info message from pipeline")

	// Parse the error from the message and add it to the local errors
	case C.GST_MESSAGE_ERROR:
		p.logger.Info("Got error message from pipeline")

		// Parse the error message
		var gerr *C.GError
		var debugInfo *C.gchar
		C.gst_message_parse_error((*C.GstMessage)(msg), (**C.GError)(unsafe.Pointer(&gerr)), (**C.gchar)(unsafe.Pointer(&debugInfo)))

		var errDetails *C.GstStructure
		C.gst_message_parse_error_details((*C.GstMessage)(msg), (**C.GstStructure)(unsafe.Pointer(&errDetails)))
		if errDetails != nil {
			defer C.gst_structure_free((*C.GstStructure)(errDetails))
			numFields := int(C.gst_structure_n_fields((*C.GstStructure)(errDetails)))
			for i := 0; i < numFields-1; i++ {
				fieldName := C.gst_structure_nth_field_name((*C.GstStructure)(errDetails), (C.guint)(i))
				fieldValue := C.gst_structure_get_value((*C.GstStructure)(errDetails), (*C.gchar)(fieldName))
				strValueDup := C.g_strdup_value_contents((*C.GValue)(fieldValue))
				p.logger.Info("Error details", "FieldName", C.GoString(fieldName), "FieldValue", C.GoString(strValueDup))
			}
		}

		if gerr != nil {
			defer C.g_error_free(gerr)
			// Convert the error message to a go error
			errMsg := C.GoString(gerr.message)
			goErr := errors.New(errMsg)

			// log the error
			p.logger.Error(goErr, "Error from streaming pipeline")

			// append the error to the pipeline
			p.errors = append(p.errors, goErr)
		}

		if debugInfo != nil {
			defer C.g_free((C.gpointer)(debugInfo))
			// Log the debug info if any
			debugStr := C.GoString((*C.gchar)(debugInfo))
			if strings.TrimSpace(debugStr) != "" {
				p.logger.Info(fmt.Sprintf("Gst debug: %s", debugStr))
			}
		}

		p.Close()

	// Record the current state of the pipeline
	case C.GST_MESSAGE_STATE_CHANGED:
		var oldState, newState C.GstState
		C.gst_message_parse_state_changed((*C.GstMessage)(msg), (*C.GstState)(unsafe.Pointer(&oldState)), (*C.GstState)(unsafe.Pointer(&newState)), nil)
		oldStateName := C.GoString(C.gst_element_state_get_name((C.GstState)(oldState)))
		newStateName := C.GoString(C.gst_element_state_get_name((C.GstState)(newState)))
		if p.currentState != newStateName {
			p.logger.Info("Got pipeline state change", "OldState", oldStateName, "NewState", newStateName)
			p.currentState = newStateName
		}

	// Messages that could be useful in the future
	case C.GST_MESSAGE_ELEMENT:
	case C.GST_MESSAGE_STREAM_STATUS:
	case C.GST_MESSAGE_BUFFERING:
	case C.GST_MESSAGE_LATENCY:
	case C.GST_MESSAGE_NEW_CLOCK:
	case C.GST_MESSAGE_ASYNC_DONE:
	case C.GST_MESSAGE_TAG:

	// To catch unhandled messages and build handlers for them
	default:
		msgTypeName := C.gst_message_type_get_name((C.GstMessageType)(msg._type))
		p.logger.Info(fmt.Sprintf("Received message with no handler: %s", C.GoString(msgTypeName)))
	}
}
