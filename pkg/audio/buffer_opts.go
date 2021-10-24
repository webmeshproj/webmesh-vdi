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

package audio

import (
	"path/filepath"

	"github.com/go-logr/logr"
	v1 "github.com/kvdi/kvdi/apis/meta/v1"
)

// BufferOpts represents options passed to NewBuffer when building
// a recording or playback pipeline. Sane defaults are provided for every
// field, but may not be suitable for all use cases.
type BufferOpts struct {
	// A Logger to log messages to, one will be created if this is nil.
	Logger logr.Logger
	// The path to the PulseAudio UNIX socket. The default server is selected
	// if omitted.
	PulseServer string
	// The format to use when streaming and writing to the pulse server.
	// Defaults to `S16LE` (signed 16-bit little-endian).
	PulseFormat string
	// The name of the device to monitor for playback on the read-buffer.
	// The default device is selected when this omitted.
	PulseMonitorName string
	// The sample rate to use on the playback monitor. Defaults to 24000.
	PulseMonitorSampleRate int
	// The number of channels to record on the playback monitor. Defaults to 2.
	PulseMonitorChannels int
	// The name of the PulseSource to write to when recording on the write-buffer.
	// This is required because an additional monitor needs to be created on the
	// mic device to allow PulseAudio to flush its buffers. Defaults to "virtmic".
	PulseMicName string
	// The path of the PulseAudio FIFO to write to when recording on the write-buffer.
	// Defaults to /var/run/kvdi/mic.fifo.
	PulseMicPath string
	// The sample rate of the pulse mic. Defaults to 16000.
	PulseMicSampleRate int
	// The number of channels on the mic. Defaults to 1.
	PulseMicChannels int
}

func (o *BufferOpts) getLogger() logr.Logger {
	return o.Logger
}

func (o *BufferOpts) getPulseFormat() string {
	if o.PulseFormat == "" {
		return "S16LE"
	}
	return o.PulseFormat
}

func (o *BufferOpts) getPulseMicRate() int {
	if o.PulseMicSampleRate == 0 {
		return 16000
	}
	return o.PulseMicSampleRate
}

func (o *BufferOpts) getPulseMicChannels() int {
	if o.PulseMicChannels == 0 {
		return 1
	}
	return o.PulseMicChannels
}

func (o *BufferOpts) getPulsePlaybackRate() int {
	if o.PulseMonitorSampleRate == 0 {
		return 24000
	}
	return o.PulseMonitorSampleRate
}

func (o *BufferOpts) getPulseMonitorChannels() int {
	if o.PulseMonitorChannels == 0 {
		return 2
	}
	return o.PulseMonitorChannels
}

func (o *BufferOpts) getPulseMonitorName() string {
	// gstreamer will attempt to use the default if this is empty
	return o.PulseMonitorName
}

func (o *BufferOpts) getPulseServer() string {
	// gstreamer will attempt to use the default if this is empty
	return o.PulseServer
}

func (o *BufferOpts) getMicName() string {
	if o.PulseMicName == "" {
		return "virtmic"
	}
	return o.PulseMicName
}

func (o *BufferOpts) getMicPath() string {
	if o.PulseMicPath == "" {
		return filepath.Join(v1.DesktopRunDir, "mic.fifo")
	}
	return o.PulseMicPath
}
