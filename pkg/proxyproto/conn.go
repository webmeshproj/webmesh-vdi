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

package proxyproto

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/go-logr/logr"
	"github.com/kvdi/kvdi/pkg/util/tlsutil"
)

// Conn represents a single connection between the app and proxy. It embeds a
// net.Conn and exports additional methods for parsing and reading the
// purpose of the request as well as tracking connection metrics.
type Conn struct {
	net.Conn
	rtype        RequestType
	rsize, wsize int64
	log          logr.Logger
}

// Dial dials the given server and initializes a new client connection for the given request
// type.
func Dial(logger logr.Logger, addr string, rtype RequestType) (*Conn, error) {
	cfg, err := tlsutil.NewClientTLSConfig()
	if err != nil {
		return nil, err
	}
	logger.Info("Dialing proxy instance", "Address", addr)
	c, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}
	pc := &Conn{
		Conn:  c,
		rtype: rtype,
		log:   logger.WithName(rtype.String()),
	}
	if err := pc.clientInit(); err != nil {
		if cerr := c.Close(); cerr != nil {
			logger.Error(cerr, "Error closing failed connection")
		}
		return nil, err
	}
	return pc, nil
}

// NewConn takes a new client connection, reads the type of request off the wire, and
// wraps it in a Conn object for processing.
func NewConn(logger logr.Logger, c net.Conn) (*Conn, error) {
	pc := &Conn{
		Conn: c,
		log:  logger.WithName(c.RemoteAddr().String()),
	}
	if err := pc.readType(); err != nil {
		if cerr := c.Close(); cerr != nil {
			pc.log.Error(cerr, "Error closing unhandled client connection")
		}
		return nil, err
	}
	return pc, nil
}

func (c *Conn) clientInit() error {
	return c.writeByte(byte(c.rtype))
}

// RequestType returns the type of the request for this connection.
func (c *Conn) RequestType() RequestType { return c.rtype }

// Read wraps the underlying Conn reader and tracks bytes read over the life of the
// connection.
func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.rsize += int64(n)
	return n, err
}

// ReadStatus reads the status header off the wire, and if not equal to RequestOK, creates
// an error with the sent message.
func (c *Conn) ReadStatus() error {
	status, err := c.readByte()
	if err != nil {
		if cerr := c.Close(); cerr != nil {
			c.log.Error(cerr, "Error closing failed connection")
		}
		return err
	}
	if RequestStatus(status) == RequestOK {
		return nil
	}
	defer c.Close()
	errMsg, err := c.readString()
	if err != nil {
		return err
	}
	return errors.New(errMsg)
}

// ReadStructure reads from the wire into the given request object. It must be a request
// object provided by this package containing a recv() method.
func (c *Conn) ReadStructure(req interface{}) error {
	if req == nil {
		return errors.New("Cannot read into nil request object")
	}
	recvr, ok := req.(interface{ recv(*Conn) error })
	if !ok {
		return errors.New("Given request object does not contain a recv() method")
	}
	return recvr.recv(c)
}

func (c *Conn) readType() error {
	rtype, err := c.readByte()
	if err != nil {
		return err
	}
	c.rtype = RequestType(rtype)
	return nil
}

func (c *Conn) readByte() (byte, error) {
	b := make([]byte, 1)
	_, err := io.ReadFull(c.Conn, b)
	if err != nil {
		return 0, err
	}
	c.rsize++
	return b[0], nil
}

// ReadString reads until the next newline sent over the connection. Request arguments
// are newline delimited strings sent immediately after the RequestType. This
// implementation satisfies the request types currently being used, but may need to be
// adapted further in the future.
func (c *Conn) readString() (string, error) {
	scanner := bufio.NewScanner(c.Conn)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	s := scanner.Text()
	c.rsize += int64(len([]byte(s)))
	return s, nil
}

// ReadInt64 is used similarly to ReadString, except it reads a signed 64-bit integer argument
// off the client connection.
func (c *Conn) readInt64() (int64, error) {
	b := make([]byte, 8)
	n, err := io.ReadFull(c.Conn, b)
	c.rsize += int64(n)
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

// Write wraps the underlying Conn writer and tracks bytes written over the life of the
// connection.
func (c *Conn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.wsize += int64(n)
	return n, err
}

func (c *Conn) writeByte(b byte) error {
	_, err := c.Conn.Write([]byte{b})
	if err != nil {
		return err
	}
	c.wsize++
	return nil
}

// WriteStructure writes the given response object to the wire. It must be a response structure declared
// in this package with a send() method.
func (c *Conn) WriteStructure(res interface{}) error {
	if res == nil {
		return errors.New("Cannot write nil response object")
	}
	sender, ok := res.(interface{ send(*Conn) error })
	if !ok {
		return errors.New("Given response object does not contain a send() method")
	}
	return sender.send(c)
}

// MustWriteStructure writes the given response object to the wire, silently logging any errors in
// the process.
func (c *Conn) mustWriteStructure(res interface{}) {
	if err := c.WriteStructure(res); err != nil {
		c.log.Error(err, "Error writing request/response object to wire")
	}
}

// WriteResponse is a convenience wrapper for writing a RequestOK followed by the given structure.
func (c *Conn) WriteResponse(res interface{}) {
	if err := c.WriteStatus(RequestOK); err != nil {
		c.log.Error(err, "Error writing response header")
		return
	}
	c.mustWriteStructure(res)
}

// WriteStatus attempts to write the given status byte to the connection.
func (c *Conn) WriteStatus(st RequestStatus) error {
	return c.writeByte(byte(st))
}

// WriteError writes the given error to the wire.
func (c *Conn) WriteError(err error) {
	if err == nil {
		c.log.Error(errors.New("Error cannot be nil"), "")
		return
	}
	c.log.Error(err, "Error during request")
	if werr := c.WriteStatus(RequestFailed); werr != nil {
		c.log.Error(werr, "Failed to write response status header")
		return
	}
	if werr := c.writeString(err.Error()); werr != nil {
		c.log.Error(werr, "Failed to send error response")
	}
}

// WriteString writes an argument to the current request. See ReadString for more information.
func (c *Conn) writeString(s string) error {
	b := append([]byte(s), []byte("\n")...)
	n, err := c.Conn.Write(b)
	c.wsize += int64(n)
	return err
}

// WriteInt64 is used similarly to WriteString to write a signed 64-bit integer argument to
// the connection.
func (c *Conn) writeInt64(num int64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(num))
	n, err := c.Conn.Write(b)
	c.wsize += int64(n)
	return err
}

// BytesRecvdCount returns the total number of bytes read on the connection so far.
func (c *Conn) BytesRecvdCount() int64 { return c.rsize }

// BytesSentCount returns the total number of bytes written to the connection so far.
func (c *Conn) BytesSentCount() int64 { return c.wsize }
