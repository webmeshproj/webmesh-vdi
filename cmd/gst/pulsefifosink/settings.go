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

package main

import (
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

var (
	defaultDeviceName = "virtmic"
)

func main() {}

var cat = gst.NewDebugCategory(
	"pulsefifosink",
	gst.DebugColorNone,
	"PulseFIFOSink Element",
)

var properties = []*glib.ParamSpec{
	glib.NewStringParam(
		"server",
		"Pulse Server",
		"Address where the PulseAudio server is listening for connections",
		nil,
		glib.ParameterReadWrite,
	),
	glib.NewStringParam(
		"device-name",
		"Device Name",
		"The name to give the virtual source device",
		&defaultDeviceName,
		glib.ParameterReadWrite,
	),
	glib.NewStringParam(
		"device-path",
		"Device Path",
		"The path on the filesystem to place the FIFO device. Defaults to a generated temporary file.",
		nil,
		glib.ParameterReadWrite,
	),
}

type settings struct {
	server     string
	deviceName string
	devicePath string
}

func defaultSettings() *settings {
	return &settings{
		deviceName: defaultDeviceName,
	}
}
