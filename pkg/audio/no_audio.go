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

package audio

import "errors"

//

// NewBuffer returns a new Buffer.
func NewBuffer(opts *BufferOpts) Buffer {
	return &unsupportedBuffer{}
}

type unsupportedBuffer struct {
}

func (b *unsupportedBuffer) Start() error {
	return errors.New("audio not supported")
}

func (b *unsupportedBuffer) RunLoop() {
}

func (b *unsupportedBuffer) Read(p []byte) (n int, err error) {
	return 0, errors.New("audio not supported")
}

func (b *unsupportedBuffer) Write(p []byte) (n int, err error) {
	return 0, errors.New("audio not supported")
}

func (b *unsupportedBuffer) Close() error {
	return errors.New("audio not supported")
}

func (b *unsupportedBuffer) IsClosed() bool {
	return true
}
