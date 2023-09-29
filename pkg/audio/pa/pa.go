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

// Package pa contains a PulseAudio C API wrapper for managing
// virtual devices on a system.
package pa

import "time"

// DeviceManagerOpts represent options to pass to the device manager.
type DeviceManagerOpts struct {
	PulseServer string
}

// SourceOpts represents options for a creating a new virtual source.
type SourceOpts struct {
	Name                 string
	Description          string
	FifoPath             string
	SampleFormat         string
	Channels, SampleRate int
}

type DeviceManager interface {
	Destroy() error
	WaitForReady(time.Duration) error
	SetDefaultSource(name string) error
	AddSink(name, description string) (Device, error)
	AddSource(opts *SourceOpts) (Device, error)
}

type Device interface{}

func NewDeviceManager(opts *DeviceManagerOpts) (DeviceManager, error) {
	return newDeviceManager(opts)
}
