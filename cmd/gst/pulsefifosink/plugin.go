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

import "C"

import (
	"unsafe"

	"github.com/tinyzimmer/go-gst/gst"
)

// The metadata for this plugin
var pluginMeta = &gst.PluginMetadata{
	MajorVersion: gst.VersionMajor,
	MinorVersion: gst.VersionMinor,
	Name:         "pulsefifosink",
	Description:  "Element for writing a stream to a FIFO source on a PulseAudio server",
	Version:      "v0.0.1",
	License:      gst.LicenseLGPL,
	Source:       "go-gst",
	Package:      "examples",
	Origin:       "https://github.com/tinyzimmer/go-gst",
	ReleaseDate:  "2021-01-04",
	// The init function is called to register elements provided by the plugin.
	Init: func(plugin *gst.Plugin) bool {
		return gst.RegisterElement(
			plugin,
			// The name of the element
			"pulsefifosink",
			// The rank of the element
			gst.RankNone,
			// The GoElement implementation for the element
			&pulsefifosink{},
			// The base subclass this element extends
			gst.ExtendsBin,
		)
	},
}

// A single method must be exported from the compiled library that provides for GStreamer
// to fetch the description and init function for this plugin. The name of the method
// must match the format gst_plugin_NAME_get_desc, where NAME is the name of the compiled
// artifact with or without the "libgst" prefix and hyphens are replaced with underscores.

//export gst_plugin_pulsefifosink_get_desc
func gst_plugin_pulsefifosink_get_desc() unsafe.Pointer { return pluginMeta.Export() }
