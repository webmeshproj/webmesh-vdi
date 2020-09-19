package gst

// #cgo pkg-config: gstreamer-1.0
// #cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
// #include <gst/gst.h>
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
func (m *Message) Native() *C.GstMessage {
	return m.msg
}

// Type returns the GstMessageType of the message.
func (m *Message) Type() C.GstMessageType {
	return m.msg._type
}

// ParseError will return a GoGError from the contents of this message. This will only work
// if the GstMessageType is GST_MESSAGE_ERROR.
func (m *Message) ParseError() *GoGError {
	var gerr *C.GError
	var debugInfo *C.gchar
	C.gst_message_parse_error((*C.GstMessage)(m.msg), (**C.GError)(unsafe.Pointer(&gerr)), (**C.gchar)(unsafe.Pointer(&debugInfo)))
	if gerr != nil {
		defer C.g_free((C.gpointer)(debugInfo))
		return &GoGError{
			gerr:     gerr,
			msg:      m.msg,
			debugStr: strings.TrimSpace(C.GoString((*C.gchar)(debugInfo))),
		}
	}
	return nil
}

// ParseStateChanged will return the old and new states as Go strings.
func (m *Message) ParseStateChanged() (oldState, newState string) {
	var gOldState, gNewState C.GstState
	C.gst_message_parse_state_changed((*C.GstMessage)(m.msg), (*C.GstState)(unsafe.Pointer(&gOldState)), (*C.GstState)(unsafe.Pointer(&gNewState)), nil)
	oldState = C.GoString(C.gst_element_state_get_name((C.GstState)(gOldState)))
	newState = C.GoString(C.gst_element_state_get_name((C.GstState)(gNewState)))
	return
}

// Unref will call `gst_message_unref` on the underlying GstMessage
func (m *Message) Unref() {
	C.gst_message_unref((*C.GstMessage)(m.msg))
}

// GoGError is a Go wrapper for a C GstError. It implements the error interface
// and provides additional function for retrieving debug strings, details, and unref-ing.
type GoGError struct {
	msg      *C.GstMessage
	gerr     *C.GError
	debugStr string
}

// Unref calls `g_error_free` on the underlying GError.
func (e *GoGError) Unref() {
	C.g_error_free(e.gerr)
}

// Error implements the error interface and returns the error message.
func (e *GoGError) Error() string {
	return C.GoString(e.gerr.message)
}

// DebugString returns any debug info alongside the error.
func (e *GoGError) DebugString() string { return e.debugStr }

// Details will returns a map of details about the error if available.
// It requires the GstMessage that produced this error to have not been unrefed yet.
func (e *GoGError) Details() map[string]string {
	var errDetails *C.GstStructure
	C.gst_message_parse_error_details((*C.GstMessage)(e.msg), (**C.GstStructure)(unsafe.Pointer(&errDetails)))
	if errDetails != nil {
		defer C.gst_structure_free((*C.GstStructure)(errDetails))
		goDetails := make(map[string]string)
		numFields := int(C.gst_structure_n_fields((*C.GstStructure)(errDetails)))
		for i := 0; i < numFields-1; i++ {
			fieldName := C.gst_structure_nth_field_name((*C.GstStructure)(errDetails), (C.guint)(i))
			fieldValue := C.gst_structure_get_value((*C.GstStructure)(errDetails), (*C.gchar)(fieldName))
			strValueDup := C.g_strdup_value_contents((*C.GValue)(fieldValue))
			goDetails[C.GoString(fieldName)] = C.GoString(strValueDup)
		}
		return goDetails
	}
	return nil
}
