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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kennygrant/sanitize"

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/audio"
	"github.com/kvdi/kvdi/pkg/audio/pa"
	"github.com/kvdi/kvdi/pkg/proxyproto"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/common"
	"github.com/kvdi/kvdi/pkg/util/errors"
)

func (p *Server) handleDisplay(conn *proxyproto.Conn) {
	addr := fmt.Sprintf("%s://%s", p.opts.DisplayProto, p.opts.DisplayAddress)
	p.log.Info(fmt.Sprintf("Received display proxy request, connecting to %s", addr))
	defer conn.Close()

	displayConn, err := net.Dial(p.opts.DisplayProto, p.opts.DisplayAddress)
	if err != nil {
		p.log.Error(err, "Failed to connect to display server")
		conn.WriteError(err)
		return
	}
	p.log.Info("Connection to display server established")

	p.log.Info("Starting display proxy")
	if err := conn.WriteStatus(proxyproto.RequestOK); err != nil {
		p.log.Error(err, "Failed to write response header")
		return
	}

	stChan := p.logConnectionMetrics("display", conn)
	defer func() { stChan <- struct{}{} }()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		if _, err := io.Copy(displayConn, conn); err != nil {
			p.log.Error(err, "Error while copying stream from client connection to display socket")
		}
	}()

	// Copy server connection to the client
	go func() {
		defer cancel()
		if _, err := io.Copy(conn, displayConn); err != nil {
			p.log.Error(err, "Error while copying stream from display socket to client connection")
		}
	}()

	p.log.Info(fmt.Sprintf("Connecting to pulse server: %s", p.opts.PulseServer))
	paDevices, err := pa.NewDeviceManager(&pa.DeviceManagerOpts{
		PulseServer: p.opts.PulseServer,
	})
	if err != nil {
		p.log.Error(err, "Failed to setup pulseaudio, playback will not function")
	}
	if err == nil {
		defer func() {
			if derr := paDevices.Destroy(); derr != nil {
				p.log.Error(derr, "Failed to cleanup device manager")
			}
		}()
		p.log.Info("Setting up audio devices")
		if err := p.setupPulseAudio(paDevices); err != nil {
			if derr := paDevices.Destroy(); derr != nil {
				p.log.Error(derr, "Failed to cleanup device manager")
			}
			p.log.Error(err, "Failed to setup pulseaudio, playback will not function")
		}
	}

	// block until the context is finished
	for range ctx.Done() {
	}
}

func (p *Server) handleAudio(conn *proxyproto.Conn) {
	p.log.Info("Received audio proxy request, setting up pulseaudio/g-streamer")
	defer conn.Close()

	p.log.Info("Starting audio buffer")
	// Create a new audio buffer
	audioBuffer := audio.NewBuffer(&audio.BufferOpts{
		Logger:                 p.log,
		PulseServer:            p.opts.PulseServer,
		PulseMonitorSampleRate: p.opts.PlaybackSampleRate,
		PulseMonitorName:       p.opts.PlaybackDeviceName,
		PulseMicName:           p.opts.RecordingDeviceName,
		PulseMicPath:           p.opts.RecordingDevicePath,
	})

	// Start the audio buffer
	if err := audioBuffer.Start(); err != nil {
		conn.WriteError(err)
		return
	}

	p.log.Info("Starting audio proxy")
	if err := conn.WriteStatus(proxyproto.RequestOK); err != nil {
		p.log.Error(err, "Failed to write response header")
		return
	}

	stChan := p.logConnectionMetrics("audio", conn)
	defer func() { stChan <- struct{}{} }()

	// Copy audio playback data to the connection
	go func() {
		defer audioBuffer.Close()
		if _, err := io.Copy(conn, audioBuffer); err != nil {
			if !errors.IsBrokenPipeError(err) {
				p.log.Error(err, "Error while copying from audio stream to websocket connection")
			}
		}
	}()

	// Copy any received recording data to the buffer
	go func() {
		defer audioBuffer.Close()
		if _, err := io.Copy(audioBuffer, conn); err != nil {
			if !errors.IsBrokenPipeError(err) {
				p.log.Error(err, "Error while copying from websocket connection to audio buffer")
			}
		}
	}()

	// Block on the audio pipeline
	audioBuffer.RunLoop()
	p.log.Info("Audio stream proxy ended")
}

func (p *Server) setupPulseAudio(manager pa.DeviceManager) error {
	if err := manager.WaitForReady(time.Second * 2); err != nil {
		return err
	}
	if _, err := manager.AddSink(p.opts.PlaybackDeviceName, p.opts.PlaybackDeviceDescription); err != nil {
		return err
	}

	if _, err := manager.AddSource(&pa.SourceOpts{
		Name:         p.opts.RecordingDeviceName,
		Description:  p.opts.RecordingDeviceDescription,
		FifoPath:     p.opts.RecordingDevicePath,
		SampleFormat: p.opts.RecordingDeviceFormat,
		SampleRate:   p.opts.RecordingDeviceSampleRate,
		Channels:     p.opts.RecordingDeviceChannels,
	}); err != nil {
		return err
	}

	if err := manager.SetDefaultSource(p.opts.RecordingDeviceName); err != nil {
		return err
	}
	return nil
}

func (p *Server) handleStat(conn *proxyproto.Conn) {
	defer conn.Close()

	req := &proxyproto.FStatRequest{}
	if err := conn.ReadStructure(req); err != nil {
		p.log.Error(err, "Could not read stat request from client")
		conn.WriteError(err)
		return
	}
	p.log.Info(req.String())

	path, err := getLocalPathFromRequest(req.Path)
	if err != nil {
		p.log.Error(err, "Could not retrieve path from request")
		conn.WriteError(err)
		return
	}
	finfo, err := os.Stat(path)
	if err != nil {
		conn.WriteError(err)
		return
	}
	resp := &types.StatDesktopFileResponse{
		Stat: &types.FileStat{
			Name:        finfo.Name(),
			IsDirectory: finfo.IsDir(),
		},
	}
	if finfo.IsDir() {
		resp.Stat.Contents = make([]*types.FileStat, 0)
		files, err := os.ReadDir(path)
		if err != nil {
			conn.WriteError(err)
			return
		}
		for _, file := range files {
			var size int64
			if !file.IsDir() {
				// Get the size
				finfo, err := file.Info()
				if err == nil {
					size = finfo.Size()
				}
			}
			resp.Stat.Contents = append(resp.Stat.Contents, &types.FileStat{
				Name:        file.Name(),
				IsDirectory: file.IsDir(),
				Size:        size,
			})
		}

	} else {
		resp.Stat.Size = finfo.Size()
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		p.log.Error(err, "Failed to marshal response")
		conn.WriteError(err)
		return
	}

	if err := conn.WriteStatus(proxyproto.RequestOK); err != nil {
		p.log.Error(err, "Failed to write response status header")
		return
	}

	if _, err := conn.Write(out); err != nil {
		p.log.Error(err, "Failed to copy response to client")
	}
}

func (p *Server) handleGet(conn *proxyproto.Conn) {
	defer conn.Close()

	req := &proxyproto.FGetRequest{}
	if err := conn.ReadStructure(req); err != nil {
		p.log.Error(err, "Could not read get request from client")
		conn.WriteError(err)
		return
	}
	p.log.Info(req.String())

	path, err := getLocalPathFromRequest(req.Path)
	if err != nil {
		p.log.Error(err, "Could not retrieve path from request")
		conn.WriteError(err)
		return
	}

	finfo, err := os.Stat(path)
	if err != nil {
		conn.WriteError(err)
		return
	}

	if finfo.IsDir() {
		serveDir(conn, path)
		return
	}

	serveFile(conn, finfo, path)
}

func (p *Server) handlePut(conn *proxyproto.Conn) {
	defer conn.Close()

	req := &proxyproto.FPutRequest{}
	if err := conn.ReadStructure(req); err != nil {
		p.log.Error(err, "Could not read put request from client")
		conn.WriteError(err)
		return
	}
	p.log.Info(req.String())

	uploadDir := filepath.Join(v1.DesktopHomeMntPath, "Uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		conn.WriteError(err)
		return
	}
	if err := os.Chown(uploadDir, p.opts.FSUserID, p.opts.FSUserID); err != nil {
		conn.WriteError(err)
		return
	}

	fName := sanitize.BaseName(req.Name)
	dstFile := filepath.Join(uploadDir, fName)

	f, err := os.Create(dstFile)
	if err != nil {
		conn.WriteError(err)
		return
	}
	defer f.Close()

	if _, err := io.CopyN(f, req.Body, req.Size); err != nil {
		conn.WriteError(err)
		return
	}

	if err := os.Chown(dstFile, p.opts.FSUserID, p.opts.FSUserID); err != nil {
		conn.WriteError(err)
		return
	}

	if err := conn.WriteStatus(proxyproto.RequestOK); err != nil {
		p.log.Error(err, "Error writing OK to connection")
	}
}

func serveDir(conn *proxyproto.Conn, path string) {
	tarball, err := common.TarDirectoryToTempFile(path)
	if err != nil {
		conn.WriteError(err)
		return
	}
	finfo, err := os.Stat(tarball)
	if err != nil {
		conn.WriteError(err)
		return
	}
	serveFile(conn, finfo, tarball)
}

func serveFile(conn *proxyproto.Conn, finfo os.FileInfo, path string) {
	f, err := os.Open(path)
	if err != nil {
		conn.WriteError(err)
		return
	}
	defer f.Close()

	// Get the file header
	hdr := make([]byte, 512)
	if _, err := f.Read(hdr); err != nil {
		conn.WriteError(errors.New("Failed to read header from file"))
		return
	}

	// Seek back to the start of the file (since we read the header already)
	if _, err := f.Seek(0, 0); err != nil {
		conn.WriteError(errors.New("Failed to seek to beginning of file"))
		return
	}

	// Get content type of file
	contentType := http.DetectContentType(hdr)

	conn.WriteResponse(&proxyproto.FGetResponse{
		Name: filepath.Base(finfo.Name()),
		Type: contentType,
		Size: finfo.Size(),
		Body: f,
	})
}
