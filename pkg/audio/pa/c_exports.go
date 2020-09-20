package pa

/*
#cgo pkg-config: libpulse
#cgo CFLAGS: -Wno-deprecated-declarations -g -Wall
#include <pulse/def.h>
*/
import "C"

import (
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

//export moduleIDCb
func moduleIDCb(idx C.uint32_t, chPtr unsafe.Pointer) {
	ptr := gopointer.Restore(chPtr)
	ch := ptr.(chan int)
	ch <- int(idx)
}

//export stateChanged
func stateChanged(managerPtr unsafe.Pointer) {
	ptr := gopointer.Restore(managerPtr)
	manager := ptr.(*DeviceManager)
	manager.stateChanged()
}

//export successCb
func successCb(success C.int, chPtr unsafe.Pointer) {
	ptr := gopointer.Restore(chPtr)
	ch := ptr.(chan bool)
	ch <- success != 0
}
