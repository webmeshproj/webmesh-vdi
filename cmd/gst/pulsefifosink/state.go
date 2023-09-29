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

package main

import (
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/kvdi/kvdi/pkg/audio/pa"
)

type state struct {
	running   bool
	isTmpDir  bool
	paDevices *pa.DeviceManager
	sinkPad   *gst.Pad
}

func (s *state) GetDeviceManager() *pa.DeviceManager  { return s.paDevices }
func (s *state) SetDeviceManager(d *pa.DeviceManager) { s.paDevices = d }
func (s *state) GetSinkPad() *gst.Pad                 { return s.sinkPad }
func (s *state) SetSinkPad(pad *gst.Pad)              { s.sinkPad = pad }
func (s *state) GetIsTempDir() bool                   { return s.isTmpDir }
func (s *state) SetIsTempDir(b bool)                  { s.isTmpDir = true }
