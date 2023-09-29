//go:build !audio

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

import (
	"errors"
	"time"
)

func newDeviceManager(opts *DeviceManagerOpts) (*deviceManager, error) {
	return &deviceManager{}, errors.New("not compiled with audio support")
}

type deviceManager struct{}

func (m *deviceManager) Destroy() error {
	return errors.New("audio not supported")
}

func (m *deviceManager) WaitForReady(time.Duration) error {
	return errors.New("audio not supported")
}

func (m *deviceManager) SetDefaultSource(name string) error {
	return errors.New("audio not supported")
}
func (m *deviceManager) AddSink(name, description string) (Device, error) {
	return nil, errors.New("audio not supported")
}
func (m *deviceManager) AddSource(opts *SourceOpts) (Device, error) {
	return nil, errors.New("audio not supported")
}
