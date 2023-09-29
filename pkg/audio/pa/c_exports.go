//go:build audio

/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	manager := ptr.(*deviceManager)
	manager.stateChanged()
}

//export successCb
func successCb(success C.int, chPtr unsafe.Pointer) {
	ptr := gopointer.Restore(chPtr)
	ch := ptr.(chan bool)
	ch <- success != 0
}
