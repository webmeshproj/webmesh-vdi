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

// Package client contains a client implementation for making requests against
// desktop proxy instances.
package client

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/kvdi/kvdi/pkg/proxyproto"
)

// Client is a structure used by the kvdi-app for sending traffic to and from
// the kvdi-proxy instances.
type Client struct {
	proxyAddr string
	log       logr.Logger
}

// New returns a new proxy client to send requests to the given address.
func New(logger logr.Logger, addr string) *Client {
	return &Client{proxyAddr: addr, log: logger}
}

func (p *Client) tryCloseError(c *proxyproto.Conn) {
	if cerr := c.Close(); cerr != nil {
		p.log.Error(cerr, "Error closing failed connection")
	}
}

// DisplayProxy returns a new connection for proxying a display stream.
func (p *Client) DisplayProxy() (*proxyproto.Conn, error) {
	c, err := proxyproto.Dial(p.log, p.proxyAddr, proxyproto.RequestTypeDisplay)
	if err != nil {
		return nil, err
	}
	if err := c.ReadStatus(); err != nil {
		return nil, err
	}
	return c, nil
}

// AudioProxy returns a new connection for proxying a display stream.
func (p *Client) AudioProxy() (*proxyproto.Conn, error) {
	c, err := proxyproto.Dial(p.log, p.proxyAddr, proxyproto.RequestTypeAudio)
	if err != nil {
		return nil, err
	}
	if err := c.ReadStatus(); err != nil {
		return nil, err
	}
	return c, nil
}

// StatFile will stat a path on the desktop's filesystem. The returned reader contains
// json to be presented to the requestor.
func (p *Client) StatFile(req *proxyproto.FStatRequest) (io.ReadCloser, error) {
	c, err := proxyproto.Dial(p.log, p.proxyAddr, proxyproto.RequestTypeFStat)
	if err != nil {
		return nil, err
	}
	if err := c.WriteStructure(req); err != nil {
		p.tryCloseError(c)
		return nil, err
	}
	if err := c.ReadStatus(); err != nil {
		return nil, err
	}
	return c, nil
}

// GetFile will retrieve a file on the desktop's filesystem.
func (p *Client) GetFile(req *proxyproto.FGetRequest) (*proxyproto.FGetResponse, error) {
	c, err := proxyproto.Dial(p.log, p.proxyAddr, proxyproto.RequestTypeFGet)
	if err != nil {
		return nil, err
	}
	if err := c.WriteStructure(req); err != nil {
		p.tryCloseError(c)
		return nil, err
	}
	if err := c.ReadStatus(); err != nil {
		return nil, err
	}
	res := &proxyproto.FGetResponse{}
	if err := c.ReadStructure(res); err != nil {
		p.tryCloseError(c)
		return nil, err
	}
	return res, nil
}

// PutFile will send a file to the desktop's filesystem.
func (p *Client) PutFile(req *proxyproto.FPutRequest) error {
	c, err := proxyproto.Dial(p.log, p.proxyAddr, proxyproto.RequestTypeFPut)
	if err != nil {
		return err
	}
	errors := make(chan error)
	// We might get an error back before we manage to send the whole request,
	// so instead of blocking on the send, block on reading back the response.
	go func() { errors <- c.WriteStructure(req) }()
	if err := c.ReadStatus(); err != nil {
		return err
	}
	// Blocks until the send is complete
	err = <-errors
	if err != nil {
		p.tryCloseError(c)
		return err
	}
	return c.Close()
}
