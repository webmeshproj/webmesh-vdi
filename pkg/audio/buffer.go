// Package audio contains utilities for streaming audio from a desktop to
// a websocket client. It is used by the kvdi-proxy to provide an audio stream
// to the web UI.
package audio

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

// Buffer provides a ReadWriteCloser for proxying audio data to
// and from a websocket connection.
type Buffer struct {
	logger               logr.Logger
	pbkPipeline          *PlaybackPipeline
	recPipeline          *RecordingPipeline
	channels, sampleRate int
	userID               string
	closed               bool
}

var _ io.ReadWriteCloser = &Buffer{}

// NewBuffer returns a new Buffer.
func NewBuffer(logger logr.Logger, userID string) *Buffer {
	return &Buffer{
		userID:     userID,
		logger:     logger.WithName("audio_buffer"),
		channels:   2,
		sampleRate: 24000,
	}
}

// SetChannels sets the number of channels to record from gstreamer. When this method is not called
// the value defaults to 2 (stereo).
func (a *Buffer) SetChannels(c int) { a.channels = c }

// SetSampleRate sets the sample rate to use when recording from gstreamer. When this method is not called
// the value defaults to 24000.
func (a *Buffer) SetSampleRate(r int) { a.sampleRate = r }

// Start starts the gstreamer processes
func (a *Buffer) Start() error {
	var err error

	a.pbkPipeline, err = NewPlaybackPipeline(
		a.logger.WithName("gst_playback"),
		&PlaybackPipelineOpts{
			PulseServer:    fmt.Sprintf("/run/user/%s/pulse/native", a.userID),
			DeviceName:     "kvdi.monitor",
			SourceFormat:   "S16LE",
			SourceRate:     a.sampleRate,
			SourceChannels: a.channels,
		})
	if err != nil {
		return err
	}
	a.recPipeline, err = NewRecordingPipeline(
		a.logger.WithName("gst_recorder"),
		&RecordingPipelineOpts{
			DeviceFifo:     filepath.Join(v1.DesktopRunDir, "mic.fifo"),
			DeviceFormat:   "S16LE",
			DeviceRate:     16000,
			DeviceChannels: 1,
		},
	)
	if err != nil {
		return err
	}

	// Start the playback device
	if err := a.pbkPipeline.Start(); err != nil {
		return err
	}

	// Start the recording device
	if err := a.recPipeline.Start(); err != nil {
		return err
	}

	return nil
}

// Wait will watch the gstreamer pipelines and return once one or both of them has
// stopped.
func (a *Buffer) Wait() error {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		if a.pbkPipeline.IsRunning() && a.recPipeline.IsRunning() {
			continue
		}
		if len(a.pbkPipeline.Errors()) > 0 {
			return errors.New("Errors occurred on the playback pipeline")
		}
		if len(a.recPipeline.Errors()) > 0 {
			return errors.New("Errors occurred on the recording pipeline")
		}
		break
	}
	return nil
}

// Errors returns any errors that ocurred during the streaming process.
func (a *Buffer) Errors() []error {
	errs := make([]error, 0)
	if pbkErrs := a.pbkPipeline.Errors(); pbkErrs != nil {
		errs = append(errs, pbkErrs...)
	}
	if recErrs := a.recPipeline.Errors(); recErrs != nil {
		errs = append(errs, recErrs...)
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Read implements ReadCloser and returns data from the audio buffer.
func (a *Buffer) Read(p []byte) (int, error) { return a.pbkPipeline.Read(p) }

// Write implements a WriteCloser and writes data to the audio buffer.
func (a *Buffer) Write(p []byte) (int, error) { return a.recPipeline.Write(p) }

// IsClosed returns true if the buffer is closed.
func (a *Buffer) IsClosed() bool {
	return a.closed
}

// Close kills the gstreamer processes and unloads pa modules.
func (a *Buffer) Close() error {
	if !a.IsClosed() {
		if err := a.pbkPipeline.Close(); err != nil {
			return err
		}
		if err := a.recPipeline.Close(); err != nil {
			return err
		}
		a.closed = true
	}
	return nil
}
