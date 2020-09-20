package gst

/*
#cgo pkg-config: gstreamer-1.0
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <gst/gst.h>
*/
import "C"
import (
	"strings"
	"unsafe"
)

// Message is a Go wrapper around a GstMessage. It provides convenience methods for
// unref-ing and parsing the underlying messages.
type Message struct {
	msg *C.GstMessage
}

// NewMessage returns a new Message from the given GstMessage.
func NewMessage(msg *C.GstMessage) *Message { return &Message{msg: msg} }

// Native returns the underlying GstMessage object.
func (m *Message) native() *C.GstMessage {
	return m.msg
}

// Type returns the GstMessageType of the message.
func (m *Message) Type() C.GstMessageType {
	return m.msg._type
}

// TypeName returns a Go string of the GstMessageType name.
func (m *Message) TypeName() string {
	return C.GoString(C.gst_message_type_get_name((C.GstMessageType)(m.Type())))
}

// getStructure returns the GstStructure in this message, using the type of the message
// to determine the method to use.
func (m *Message) getStructure() map[string]string {
	var st *C.GstStructure

	switch m.Type() {
	case C.GST_MESSAGE_ERROR:
		C.gst_message_parse_error_details((*C.GstMessage)(m.native()), (**C.GstStructure)(unsafe.Pointer(&st)))
	case C.GST_MESSAGE_INFO:
		C.gst_message_parse_info_details((*C.GstMessage)(m.native()), (**C.GstStructure)(unsafe.Pointer(&st)))
	case C.GST_MESSAGE_WARNING:
		C.gst_message_parse_warning_details((*C.GstMessage)(m.native()), (**C.GstStructure)(unsafe.Pointer(&st)))
	}

	// if no structure was returned, immediately return nil
	if st == nil {
		return nil
	}

	// The returned structure must not be freed. Applies to all methods.
	// https://gstreamer.freedesktop.org/documentation/gstreamer/gstmessage.html#gst_message_parse_error_details
	return structureToGoMap(st)
}

// parseToError returns a new GoGError from this message instance. There are multiple
// message types that parse to this interface.
func (m *Message) parseToError() *GoGError {
	var gerr *C.GError
	var debugInfo *C.gchar

	switch m.Type() {
	case C.GST_MESSAGE_ERROR:
		C.gst_message_parse_error((*C.GstMessage)(m.native()), (**C.GError)(unsafe.Pointer(&gerr)), (**C.gchar)(unsafe.Pointer(&debugInfo)))
	case C.GST_MESSAGE_INFO:
		C.gst_message_parse_info((*C.GstMessage)(m.native()), (**C.GError)(unsafe.Pointer(&gerr)), (**C.gchar)(unsafe.Pointer(&debugInfo)))
	case C.GST_MESSAGE_WARNING:
		C.gst_message_parse_warning((*C.GstMessage)(m.native()), (**C.GError)(unsafe.Pointer(&gerr)), (**C.gchar)(unsafe.Pointer(&debugInfo)))
	}

	// if error was nil return immediately
	if gerr == nil {
		return nil
	}

	// cleanup the C error immediately and let the garbage collector
	// take over from here.
	defer C.g_error_free((*C.GError)(gerr))
	defer C.g_free((C.gpointer)(debugInfo))
	return &GoGError{
		errMsg:   C.GoString(gerr.message),
		details:  m.getStructure(),
		debugStr: strings.TrimSpace(C.GoString((*C.gchar)(debugInfo))),
	}
}

// ParseInfo is identical to ParseError. The returned types are the same. However,
// this is intended for use with GstMessageType `GST_MESSAGE_INFO`.
func (m *Message) ParseInfo() *GoGError {
	return m.parseToError()
}

// ParseWarning is identical to ParseError. The returned types are the same. However,
// this is intended for use with GstMessageType `GST_MESSAGE_WARNING`.
func (m *Message) ParseWarning() *GoGError {
	return m.parseToError()
}

// ParseError will return a GoGError from the contents of this message. This will only work
// if the GstMessageType is `GST_MESSAGE_ERROR`.
func (m *Message) ParseError() *GoGError {
	return m.parseToError()
}

// ParseStateChanged will return the old and new states as Go strings. This will only work
// if the GstMessageType is `GST_MESSAGE_STATE_CHANGED`.
func (m *Message) ParseStateChanged() (oldState, newState string) {
	var gOldState, gNewState C.GstState
	C.gst_message_parse_state_changed((*C.GstMessage)(m.native()), (*C.GstState)(unsafe.Pointer(&gOldState)), (*C.GstState)(unsafe.Pointer(&gNewState)), nil)
	oldState = C.GoString(C.gst_element_state_get_name((C.GstState)(gOldState)))
	newState = C.GoString(C.gst_element_state_get_name((C.GstState)(gNewState)))
	return
}

// Unref will call `gst_message_unref` on the underlying GstMessage
func (m *Message) Unref() {
	C.gst_message_unref((*C.GstMessage)(m.native()))
}

// GoGError is a Go wrapper for a C GError. It implements the error interface
// and provides additional functions for retrieving debug strings and details.
type GoGError struct {
	errMsg, debugStr string
	details          map[string]string
}

// Message is an alias to `Error()`. It's for clarity when this object
// is parsed from a `GST_MESSAGE_INFO` or `GST_MESSAGE_WARNING`.
func (e *GoGError) Message() string { return e.Error() }

// Error implements the error interface and returns the error message.
func (e *GoGError) Error() string { return e.errMsg }

// DebugString returns any debug info alongside the error.
func (e *GoGError) DebugString() string { return e.debugStr }

// Details contains additional metadata about the error if available.
func (e *GoGError) Details() map[string]string { return e.details }
