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

// Package server contains the server for handling requests against a desktop's
// proxy instance.
package server

import (
	"crypto/tls"
	"net"
	"strconv"

	"github.com/go-logr/logr"

	"github.com/tinyzimmer/kvdi/pkg/proxyproto"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
)

// Server is a structure used by the kvdi-proxy for accepting connections from
// the kvdi-app instances.
type Server struct {
	host string
	port int32
	opts *ProxyOpts
	log  logr.Logger
}

// ProxyOpts are additional options for configuring the proxy server.
type ProxyOpts struct {
	FSUserID                                           int
	DisplayAddress, DisplayProto                       string
	PulseServer                                        string
	PlaybackSampleRate                                 int
	PlaybackDeviceName, PlaybackDeviceDescription      string
	RecordingDeviceName, RecordingDeviceDescription    string
	RecordingDevicePath, RecordingDeviceFormat         string
	RecordingDeviceSampleRate, RecordingDeviceChannels int
}

// New returns a new proxy server configured to listen on the given host and
// port.
func New(logger logr.Logger, host string, port int32, opts *ProxyOpts) *Server {
	return &Server{
		host: host,
		port: port,
		opts: opts,
		log:  logger,
	}
}

// ListenAndServe listens and accepts incoming client connections and feeds them to
// the channel.
func (p *Server) ListenAndServe() error {
	tlsConfig, err := tlsutil.NewServerTLSConfig()
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(p.host, strconv.Itoa(int(p.port)))
	p.log.Info("Listening for new mTLS TCP connections", "Address", addr)
	l, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	for {
		c, err := l.Accept()
		if err != nil {
			p.log.Error(err, "Error accepting new client connection")
			continue
		}
		go p.handleConn(c)
	}
}

// Handler is a function for handling a connection server side.
type Handler func(*proxyproto.Conn)

// ServerHandler returns the server handler for this request type.
func (p *Server) handler(rt proxyproto.RequestType) Handler {
	switch rt {
	case proxyproto.RequestTypeDisplay:
		return p.handleDisplay
	case proxyproto.RequestTypeAudio:
		return p.handleAudio
	case proxyproto.RequestTypeFStat:
		return p.handleStat
	case proxyproto.RequestTypeFGet:
		return p.handleGet
	case proxyproto.RequestTypeFPut:
		return p.handlePut
	}
	return nil
}

func (p *Server) handleConn(c net.Conn) {
	pc, err := proxyproto.NewConn(p.log, c)
	if err != nil {
		p.log.Error(err, "Error initiating new client connection")
	}
	p.log.Info("Serving new request", "Type", pc.RequestType().String(), "Client", pc.Conn.RemoteAddr().String())
	hdlr := p.handler(pc.RequestType())
	if hdlr == nil {
		p.log.Info("No handler for request")
		return
	}
	hdlr(pc)
}
