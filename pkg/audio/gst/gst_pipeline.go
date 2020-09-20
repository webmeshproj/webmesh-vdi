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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	// A channel where a caller can listen for errors asynchronously.
	errCh chan error
	//  populated with any errors on the pipeline
	gerror error
	// the current state of the pipeline
	currentState string
}

// NewPipeline builds and returns a new empty Pipeline instance.
func NewPipeline(logger logr.Logger) (*Pipeline, error) {
	pipeline, err := newEmptyPipeline()
	if err != nil {
		return nil, err
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = logf.Log.WithName("gst_pipeline")
	}
	return &Pipeline{
		logger:          logger,
		pipelineElement: pipeline,
		reader:          r,
		writer:          w,
		rbuf:            bufio.NewReader(r),
		wbuf:            bufio.NewWriter(w),
		stopCh:          make(chan struct{}),
	}, nil
}

// NewPipelineFromConfig builds a new pipeline from the given PipelineConfig. The plugins provided
// in the configuration will be linked in the order they are given.
func NewPipelineFromConfig(logger logr.Logger, cfg *PipelineConfig) (pipeline *Pipeline, err error) {
	// create a new empty pipeline instance
	pipeline, err = NewPipeline(logger)
	if err != nil {
		return nil, err
	}
	// if any error happens while setting up the pipeline, immediately free it
	defer func() {
		if err != nil {
			if cerr := pipeline.Close(); cerr != nil {
				logger.Error(cerr, "Failed to close pipeline")
			}
		}
	}()

	// retrieve a list of the plugin names
	pluginNames := cfg.PluginNames()

	// build all the elements
	var elements map[string]*Element
	elements, err = NewElementMany(pluginNames...)
	if err != nil {
		return
	}

	// iterate the plugin names and add them to the pipeline
	for idx, name := range pluginNames {
		// get the current plugin and element
		currentPlugin := cfg.GetPluginByName(name)
		currentElem := elements[name]

		// If this is an internal sink, set the fd to the writer pipe
		if currentPlugin.InternalSink {
			if err = currentElem.Set("fd", int(pipeline.writer.Fd())); err != nil {
				return
			}
		}
		// If this is an internal source, set the fd to the reader pipe
		if currentPlugin.InternalSource {
			if err = currentElem.Set("fd", int(pipeline.reader.Fd())); err != nil {
				return
			}
		}

		// Iterate any data with the plugin and set it on the element
		for key, value := range currentPlugin.Data {
			if err = currentElem.Set(key, value); err != nil {
				return
			}
		}

		// Add the element to the pipeline
		if err = pipeline.binAdd(currentElem); err != nil {
			return
		}

		// If this is the first element continue
		if idx == 0 {
			continue
		}

		// get the last element in the chain
		lastPluginName := pluginNames[idx-1]
		lastElem := elements[lastPluginName]
		lastPlugin := cfg.GetPluginByName(lastPluginName)

		if lastPlugin == nil {
			// this should never happen, since only used internally,
			// but safety from panic
			continue
		}

		// If there are sink caps on the last element, do a filtered link to this one and continue
		if lastPlugin.SinkCaps != nil {
			if err = pipeline.ElementLinkFiltered(lastElem, currentElem, lastPlugin.SinkCaps); err != nil {
				return
			}
			continue
		}

		// link the last element to this element
		if err = pipeline.elementLink(lastElem, currentElem); err != nil {
			return
		}
	}

	return
}

// NewPipelineFromLaunchString returns a new GstPipeline from the given launch string. If useFdSrc or useFdSink
// are true, then the pipeline string is additionally formatted with the internal read/write buffers.
func NewPipelineFromLaunchString(logger logr.Logger, launchStr string, useFdSrc, useFdSink bool) (*Pipeline, error) {
	var pipeline *Pipeline
	var err error

	// create a new empty pipeline
	pipeline, err = NewPipeline(logger)
	if err != nil {
		return nil, err
	}
	// free the underlying GstPipeline element so we can replace it
	pipeline.freePipeline()

	// format source/sink with file descriptors of the pipeline
	if useFdSrc {
		launchStr = fmt.Sprintf("fdsrc fd=%d ! %s", int(pipeline.ReaderFd()), launchStr)
	}
	if useFdSink {
		launchStr = fmt.Sprintf("%s ! fdsink fd=%d", launchStr, int(pipeline.WriterFd()))
	}

	// replace the GstPipeline with one parsed from the string
	pipeline.pipelineElement, err = newPipelineFromString(launchStr)

	return pipeline, err
}

// native returns the pointer to the underlying GstPipeline element.
func (p *Pipeline) native() *C.GstElement { return p.pipelineElement }

// unsafe returns the unsafe.Pointer of the underlying GstPipeline element.
func (p *Pipeline) unsafe() unsafe.Pointer { return unsafe.Pointer(p.native()) }

// bin returns the GstBin for this GstPipeline.
func (p *Pipeline) bin() *C.GstBin { return (*C.GstBin)(p.unsafe()) }

// Read implements a Reader and returns data from the read buffer.
func (p *Pipeline) Read(b []byte) (int, error) {
	if p.IsClosed() {
		return 0, io.ErrClosedPipe
	}
	return p.rbuf.Read(b)
}

// ReaderFd returns the file descriptor for the read buffer.
func (p *Pipeline) ReaderFd() uintptr { return p.reader.Fd() }

// Write implements a Writer and places data in the write buffer.
func (p *Pipeline) Write(b []byte) (int, error) {
	if p.IsClosed() {
		return 0, io.ErrClosedPipe
	}
	return p.wbuf.Write(b)
}

// WriterFd returns the file descriptor for the write buffer.
func (p *Pipeline) WriterFd() uintptr { return p.writer.Fd() }

// ErrorChan returns the channel for listening for errors on the pipeline.
// This function must be called for the channel to be setup.
func (p *Pipeline) ErrorChan() chan error {
	if p.errCh == nil {
		p.errCh = make(chan error)
	}
	return p.errCh
}

// Start will start the GstPipeline. It is asynchronous so it does not need to be
// called within a goroutine, however, it is still safe to do so.
func (p *Pipeline) Start() error {
	p.logger.Info("Starting pipeline")
	go p.watchPipelineBus()
	return p.startPipeline()
}

// IsRunning will return true if the gstreamer pipeline is currently running.
func (p *Pipeline) IsRunning() bool {
	return !p.IsClosed()
}

// IsClosed will return true if Closed has been called on this Pipeline element.
func (p *Pipeline) IsClosed() bool { return p.native() == nil }

// Close implements a Closer and stops the pipeline and closes all buffers.
func (p *Pipeline) Close() error {
	p.logger.Info("Stopping pipeline")
	if err := p.stopPipeline(); err != nil {
		return err
	}
	p.logger.Info("Closing buffers")
	if err := p.reader.Close(); err != nil {
		return err
	}
	if err := p.writer.Close(); err != nil {
		return err
	}
	return nil
}

// startPipeline will set the GstPipeline to the PLAYING state.
func (p *Pipeline) startPipeline() error {
	stateRet := C.gst_element_set_state((*C.GstElement)(p.native()), C.GST_STATE_PLAYING)
	if stateRet == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("Failed to start pipeline")
	}
	return nil
}

// freePipeline will free the GstPipeline and signal local goroutines to stop.
func (p *Pipeline) freePipeline() {
	C.gst_object_unref((C.gpointer)(p.native()))
	p.pipelineElement = nil
}

// stopPipeline signals the GstPipeline to stop
func (p *Pipeline) stopPipeline() error {
	stateRet := C.gst_element_set_state((*C.GstElement)(p.native()), C.GST_STATE_NULL)
	if stateRet == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("Failed to stop pipeline")
	}
	p.logger.Info("Freeing pipeline resources")
	p.freePipeline()
	// currently the bus watch will return nil if the pipeline is destroyed
	// allowed for te goroutine to catch the stop on the select. this should be
	// refactored.
	p.stopCh <- struct{}{}
	return nil
}

// Errors returns any error that happened during the pipeline.
func (p *Pipeline) Error() error { return p.gerror }

// BinAddMany is a go implementation of `gst_bin_add_many` to compensate for the inability
// to use variadic functions in cgo.
func (p *Pipeline) BinAddMany(elems ...*Element) error {
	for _, elem := range elems {
		if err := p.binAdd(elem); err != nil {
			return err
		}
	}
	return nil
}

// binAdd wraps `gst_bin_add`.
func (p *Pipeline) binAdd(elem *Element) error {
	if ok := C.gst_bin_add((*C.GstBin)(p.bin()), (*C.GstElement)(elem.native())); !gobool(ok) {
		return fmt.Errorf("Failed to add element to pipeline: %s", elem.Name())
	}
	return nil
}

// ElementLinkMany is a go implementation of `gst_element_link_many` to compensate for
// no variadic functions in cgo.
func (p *Pipeline) ElementLinkMany(elems ...*Element) error {
	for idx, elem := range elems {
		if idx == 0 {
			// skip the first one as the loop always links previous to current
			continue
		}
		if err := p.elementLink(elems[idx-1], elem); err != nil {
			return err
		}
	}
	return nil
}

// elementLink wraps `gst_element_link`.
func (p *Pipeline) elementLink(beforeElem, afterElem *Element) error {
	if ok := C.gst_element_link((*C.GstElement)(beforeElem.native()), (*C.GstElement)(afterElem.native())); !gobool(ok) {
		return fmt.Errorf("Failed to link %s to %s", beforeElem.Name(), afterElem.Name())
	}
	return nil
}

// ElementLinkFiltered is a convenience wrapper around `gst_element_link_filtered` for linking caps
// between two elements.
func (p *Pipeline) ElementLinkFiltered(beforeElem, afterElem *Element, caps *Caps) error {
	gstCaps, err := caps.ToGstCaps()
	if err != nil {
		return err
	}
	if ok := C.gst_element_link_filtered((*C.GstElement)(beforeElem.native()), (*C.GstElement)(afterElem.native()), (*C.GstCaps)(gstCaps)); !gobool(ok) {
		return fmt.Errorf("Failed to link %s to %s with provider caps", beforeElem.Name(), afterElem.Name())
	}
	return nil
}

// GetBus returns the GstBus for retrieving messages from the pipeline.
func (p *Pipeline) GetBus() (*Bus, error) {
	bus := C.gst_element_get_bus((*C.GstElement)(p.native()))
	if bus == nil {
		return nil, errors.New("Could not retrieve bus from pipeline")
	}
	return &Bus{bus: bus, pipeline: p}, nil
}

// watchPipelineBus pops messages on the bus and passes them to the message handler.
func (p *Pipeline) watchPipelineBus() {
	bus, err := p.GetBus()
	if err != nil {
		p.logger.Error(err, "Stopping message queue")
		return
	}
	defer bus.Unref()
	for {
		select {
		case <-p.stopCh:
			p.logger.Info("Pipeline element has been stopped")
			return
		default:
			// Extra check that the pipeline still exists
			if p.native() == nil {
				p.logger.Info("Pipeline element has been stopped")
				return
			}
			msg := bus.BlockPopMessage()
			if msg == nil {
				continue
			}
			p.handleMessage(msg)
		}
	}
}

// logInfoOrWarning will parse and log the given info or warning message.
func (p *Pipeline) logInfoOrWarning(msg *Message) {
	p.logger.Info(fmt.Sprintf("%s from pipeline", strings.ToUpper(msg.TypeName())))

	var info *GoGError
	// These functions all do the same thing but defined here explicitly for
	// clarity.
	switch msg.Type() {
	case C.GST_MESSAGE_INFO:
		info = msg.ParseInfo()
	case C.GST_MESSAGE_WARNING:
		info = msg.ParseWarning()
	default:
		info = msg.ParseError()
	}

	// If we were unable to parse at all, return
	if info == nil {
		return
	}

	p.logger.Info(info.Message())
	if debugStr := info.DebugString(); debugStr != "" {
		p.logger.Info(fmt.Sprintf("GST Debug: %s", debugStr))
	}
	for key, value := range info.Details() {
		p.logger.Info("Details", "Type", strings.ToUpper(msg.TypeName()), "Field", key, "Value", value)
	}
}

// handleMessage handles a GstMessage on the pipeline bus.
func (p *Pipeline) handleMessage(msg *Message) {
	// unref the message after processing
	defer msg.Unref()

	switch msg.Type() {

	case C.GST_MESSAGE_STREAM_START:
		p.logger.Info("Stream has started, audio data is available on the buffer")

	case C.GST_MESSAGE_EOS:
		p.logger.Info("Stream has ended, closing pipeline")
		p.Close()

	// Parse the info message and log it to the console
	case C.GST_MESSAGE_INFO:
		p.logInfoOrWarning(msg)

	// Parse the warning message and the log it to the console
	case C.GST_MESSAGE_WARNING:
		p.logInfoOrWarning(msg)

	// Parse the error from the message and add it to the local errors
	case C.GST_MESSAGE_ERROR:
		p.logger.Info("Got error message from pipeline")
		gError := msg.ParseError()
		if gError == nil {
			return
		}
		p.logger.Error(gError, "Error from pipeline")
		if debugStr := gError.DebugString(); debugStr != "" {
			p.logger.Info(fmt.Sprintf("GST Debug: %s", debugStr))
		}
		for key, value := range gError.Details() {
			p.logger.Info("Error details", "Field", key, "Value", value)
		}
		// because of this, techincally there can only ever be one error.
		// should look into this more.
		p.gerror = gError
		if p.errCh != nil {
			p.errCh <- gError
		}
		p.logger.Info("Stopping the pipeline and freeing resources")
		p.Close()
		return

	// Record the current state of the pipeline
	case C.GST_MESSAGE_STATE_CHANGED:
		oldState, newState := msg.ParseStateChanged()
		if p.currentState != newState {
			p.logger.Info("Got pipeline state change", "OldState", oldState, "NewState", newState)
			p.currentState = newState
		}

	// Messages that could be useful in the future
	case C.GST_MESSAGE_ELEMENT:
	case C.GST_MESSAGE_STREAM_STATUS:
	case C.GST_MESSAGE_BUFFERING:
	case C.GST_MESSAGE_LATENCY:
	case C.GST_MESSAGE_NEW_CLOCK:
	case C.GST_MESSAGE_ASYNC_DONE:
	case C.GST_MESSAGE_TAG:

	// For development aid to catch additional unhandled messages that might be useful
	default:
		p.logger.Info("Received message with no handler", "MessageType", msg.TypeName())
	}
}
