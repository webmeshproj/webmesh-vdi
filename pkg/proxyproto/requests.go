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

// Package proxyproto contains the core types for interactions between the kvdi API
// and desktop proxy instances. Subpackages contain client and server implementations.
//
// The original implementation involved the proxies serving full HTTP/Websocket servers
// and the API simply proxying requests back to them. The purpose of proxyproto is to
// have more control over the interaction and reduce overhead as much as possible. For
// example, an audio proxy just involves one byte being transmitted back and forth (after
// TLS negotiation) before the stream can start, instead of the overhead of a second HTTP
// request and Websocket upgrade.
//
// I'm obviously not going to write a full RFC for this protocol, but it should definitely
// at least have its capabilities documented further.
package proxyproto

import (
	"fmt"
	"io"
)

// RequestType represents the type of request being made from a client to a proxy.
type RequestType byte

const (
	_ RequestType = iota
	// RequestTypeDisplay is a request for an interactive display feed.
	RequestTypeDisplay
	// RequestTypeAudio is a request for a bidirectional audio feed.
	RequestTypeAudio
	// RequestTypeFStat is a request for stat information for a file in the system.
	RequestTypeFStat
	// RequestTypeFGet is a request to retrieve a file from the system.
	RequestTypeFGet
	// RequestTypeFPut is a request to put a file on the system.
	RequestTypeFPut
)

// RequestStatus represents the non-wire related status of a request.
type RequestStatus byte

const (
	_ RequestStatus = iota
	// RequestOK means the request succeeded and everything following on the wire is the response
	RequestOK
	// RequestFailed means the request failed and a string representing the error is next on the wire.
	RequestFailed
)

func (r RequestType) String() string {
	switch r {
	case RequestTypeDisplay:
		return "display"
	case RequestTypeAudio:
		return "audio"
	case RequestTypeFStat:
		return "stat-file"
	case RequestTypeFGet:
		return "get-file"
	case RequestTypeFPut:
		return "put-file"
	default:
		return "unknown"
	}
}

// FStatRequest contains the parameters for sending a stat request to a proxy.
type FStatRequest struct {
	Path string
}

func (f *FStatRequest) String() string {
	return fmt.Sprintf("FStat { Path: $HOME/%s }", f.Path)
}

func (f *FStatRequest) send(c *Conn) (err error) {
	return c.writeString(f.Path)
}

func (f *FStatRequest) recv(c *Conn) (err error) {
	f.Path, err = c.readString()
	return err
}

// FGetRequest contains the parameters for sending a get file request to a proxy.
type FGetRequest struct {
	Path string
}

func (f *FGetRequest) String() string {
	return fmt.Sprintf("FGet { Path: $HOME/%s }", f.Path)
}

func (f *FGetRequest) send(c *Conn) (err error) {
	return c.writeString(f.Path)
}

func (f *FGetRequest) recv(c *Conn) (err error) {
	f.Path, err = c.readString()
	return err
}

// FGetResponse contains the response to a get file request.
type FGetResponse struct {
	Name string
	Type string
	Size int64
	Body io.ReadCloser
}

func (f *FGetResponse) send(c *Conn) (err error) {
	defer f.Body.Close()
	if err = c.writeString(f.Name); err != nil {
		return
	}
	if err = c.writeString(f.Type); err != nil {
		return
	}
	if err = c.writeInt64(f.Size); err != nil {
		return
	}
	_, err = io.Copy(c, f.Body)
	return
}

func (f *FGetResponse) recv(c *Conn) (err error) {
	if f.Name, err = c.readString(); err != nil {
		return
	}
	if f.Type, err = c.readString(); err != nil {
		return
	}
	if f.Size, err = c.readInt64(); err != nil {
		return
	}
	f.Body = c
	return
}

// FPutRequest contains the parameters for uploading a file to the desktop.
type FPutRequest struct {
	Name string
	Size int64
	Body io.ReadCloser
}

func (f *FPutRequest) String() string {
	return fmt.Sprintf("FPut { Name: %s, Size: %d }", f.Name, f.Size)
}

func (f *FPutRequest) send(c *Conn) (err error) {
	defer f.Body.Close()
	if err = c.writeString(f.Name); err != nil {
		return
	}
	if err = c.writeInt64(f.Size); err != nil {
		return
	}
	_, err = io.Copy(c, f.Body)
	return err
}

func (f *FPutRequest) recv(c *Conn) (err error) {
	if f.Name, err = c.readString(); err != nil {
		return
	}
	if f.Size, err = c.readInt64(); err != nil {
		return
	}
	f.Body = c
	return
}
