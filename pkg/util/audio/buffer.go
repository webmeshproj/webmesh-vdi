// Package audio contains utilities for streaming audio from a desktop to
// a websocket client. It is used by the novnc-proxy to provide an audio stream
// to the web UI.
package audio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/go-logr/logr"
)

// BufferType represents a type of audio buffer to spawn
type BufferType string

const (
	// BufferTypeGST is a buffer that uses a gstreamer pipeline to encode
	// PCM data from pulseaudio to the selected codec.
	BufferTypeGST BufferType = "gst"
	// BufferTypePARec is a buffer that uses parec to output raw PCM data
	// from pulseaudio and pipes to the selected encoder.
	BufferTypePARec BufferType = "parec"
)

// Codec represents the encoder to use to process the raw PCM data.
type Codec string

const (
	// CodecOpus encodes the audio with opus and wraps it in a webm container.
	CodecOpus Codec = "opus"
	// CodecVorbis encodes the audio with vorbis and wraps it in an ogg container.
	CodecVorbis Codec = "vorbis"
	// CodecMP3 encodes the audio with lame and returns it in MP3 format.
	CodecMP3 Codec = "mp3"
)

// Buffer provides a Reader interface for proxying audio data to a websocket
// connection
type Buffer struct {
	cmd        *exec.Cmd
	buffer     io.ReadCloser
	stderr     bytes.Buffer
	closed     bool
	logger     logr.Logger
	userID     string
	bufferType BufferType
}

var _ io.ReadCloser = &Buffer{}

// NewBuffer returns a new Buffer.
func NewBuffer(logger logr.Logger, userID string, bufferType BufferType) *Buffer {
	return &Buffer{
		userID:     userID,
		logger:     logger,
		bufferType: bufferType,
	}
}

func (a *Buffer) buildGSTPipeline(codec Codec) string {
	pipeline := fmt.Sprintf(
		"sudo -u audioproxy gst-launch-1.0 -q pulsesrc server=/run/user/%s/pulse/native ! audio/x-raw, channels=2, rate=24000",
		a.userID,
	)
	switch codec {
	case CodecVorbis:
		pipeline = fmt.Sprintf("%s ! vorbisenc ! oggmux", pipeline)
	case CodecOpus:
		pipeline = fmt.Sprintf("%s ! cutter ! opusenc ! webmmux", pipeline)
	case CodecMP3:
		pipeline = fmt.Sprintf("%s ! lamemp3enc", pipeline)
	default:
		a.logger.Info(fmt.Sprintf("Invalid codec for gst pipeline %s, defaulting to opus/webm", codec))
		pipeline = fmt.Sprintf("%s ! cutter ! opusenc ! webmmux", pipeline)
	}

	return fmt.Sprintf("%s ! fdsink fd=1", pipeline)
}

func (a *Buffer) buildPaRecPipeline(codec Codec) string {
	pipeline := fmt.Sprintf("sudo -u audioproxy parec -s /run/user/%s/pulse/native", a.userID)

	switch codec {
	case CodecVorbis:
		pipeline = fmt.Sprintf("%s | oggenc -b 192 -o - --raw -", pipeline)
	case CodecMP3:
		pipeline = fmt.Sprintf("%s | lame -r -V0 -", pipeline)
	default:
		a.logger.Info(fmt.Sprintf("Invalid codec for parec pipeline %s, defaulting to ogg/vorbis", codec))
		pipeline = fmt.Sprintf("%s | oggenc -b 192 -o - --raw -", pipeline)
	}
	return pipeline
}

// Start starts the gstreamer process
func (a *Buffer) Start(codec Codec) error {

	var pipeline string
	if a.bufferType == BufferTypeGST {
		pipeline = a.buildGSTPipeline(codec)
	} else if a.bufferType == BufferTypePARec {
		pipeline = a.buildPaRecPipeline(codec)
	}

	a.logger.Info(fmt.Sprintf("Running command: %s", pipeline))

	a.cmd = exec.Command("/bin/sh", "-c", pipeline)

	var err error

	a.buffer, err = a.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errPipe, err := a.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		if _, err := io.Copy(&a.stderr, errPipe); err != nil {
			a.logger.Error(err, "Erroring reading stderr from recorder process")
		}
	}()

	if err := a.cmd.Start(); err != nil {
		return err
	}

	return nil
}

// Wait will join to the streaming process and block until its finished,
// returning an error if the process exits non-zero.
func (a *Buffer) Wait() error {
	return a.cmd.Wait()
}

// Error returns any errors that ocurred during the streaming process.
func (a *Buffer) Error() error {
	if a.cmd.ProcessState.Exited() {
		if a.cmd.ProcessState.ExitCode() != 0 {
			return errors.New(a.stderr.String())
		}
	}
	return nil
}

// Stderr returns any output from stderr on the gstreamer process.
func (a *Buffer) Stderr() string { return a.stderr.String() }

// Read implements ReadCloser and returns data from the audio buffer.
func (a *Buffer) Read(p []byte) (int, error) {
	return a.buffer.Read(p)
}

// IsClosed returns true if the buffer is closed.
func (a *Buffer) IsClosed() bool { return a.closed }

// Close kills the gstreamer process.
func (a *Buffer) Close() error {
	if !a.IsClosed() {
		a.closed = true
		return a.cmd.Process.Kill()
	}
	return nil
}
