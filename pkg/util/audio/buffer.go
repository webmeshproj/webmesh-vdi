// Package audio contains utilities for streaming audio from a desktop to
// a websocket client. It is used by the kvdi-proxy to provide an audio stream
// to the web UI.
package audio

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"

	"github.com/tinyzimmer/kvdi/pkg/util/audio/gst"
	"github.com/tinyzimmer/kvdi/pkg/util/audio/pa"
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
	// CodecRaw uses raw PCM data with the configured sample rate
	CodecRaw Codec = "raw"
)

// Buffer provides a Reader interface for proxying audio data to a websocket
// connection
type Buffer struct {
	logger                   logr.Logger
	deviceManager            *pa.DeviceManager
	pbkPipeline, recPipeline *gst.Pipeline
	channels, sampleRate     int
	userID                   string
	closed                   bool
}

var _ io.ReadCloser = &Buffer{}

// NewBuffer returns a new Buffer.
func NewBuffer(logger logr.Logger, userID string) *Buffer {
	return &Buffer{
		deviceManager: pa.NewDeviceManager(logger.WithName("pa_devices"), userID),
		userID:        userID,
		logger:        logger.WithName("audio_buffer"),
		channels:      2,
		sampleRate:    24000,
	}
}

// buildPlaybackPipeline builds a GST pipeline for recording data from the dummy monitor
// and making it available on the io.Reader interface.
func (a *Buffer) buildPlaybackPipeline(codec Codec) *gst.Pipeline {
	pipeline := gst.NewPipeline(a.userID, a.logger.WithName("gst_playback")).
		WithPulseSrc(a.userID, "kvdi.monitor", a.channels, a.sampleRate)

	switch codec {
	case CodecVorbis:
		pipeline = pipeline.WithVorbisEncode().WithOggMux()
	case CodecOpus:
		pipeline = pipeline.WithCutter().WithOpusEncode().WithWebmMux()
	case CodecMP3:
		pipeline = pipeline.WithLameEncode()
	default:
		a.logger.Info(fmt.Sprintf("Invalid codec for gst pipeline %s, defaulting to opus/webm", codec))
		pipeline = pipeline.WithCutter().WithOpusEncode().WithWebmMux()
	}

	return pipeline.WithFdSink(1)
}

// buildRecordingPipeline builds a GST pipeline for receiving data from the Write interface
// and writing it to the source on the pipeline.
func (a *Buffer) buildRecordingPipeline() *gst.Pipeline {
	recPipeline := gst.NewPipeline(a.userID, a.logger.WithName("gst_recorder")).
		WithFdSrc(0, false).
		WithDecodeBin().
		WithAudioConvert().WithAudioResample().WithRawCaps("s16le", 1, 16000).
		WithFileSink(a.getMicFifo(), true)
	return recPipeline
}

func (a *Buffer) setupDevices() error {
	if err := a.deviceManager.AddSink("kvdi", "kvdi-playback"); err != nil {
		return err
	}

	if err := a.deviceManager.AddSource("virtmic", "kvdi-microphone", a.getMicFifo(), "s16le", 1, 16000); err != nil {
		return err
	}

	return a.deviceManager.SetDefaultSource("virtmic")
}

func (a *Buffer) getMicFifo() string {
	return fmt.Sprintf("/run/user/%s/pulse/mic.fifo", a.userID)
}

// SetChannels sets the number of channels to record from gstreamer. When this method is not called
// the value defaults to 2 (stereo).
func (a *Buffer) SetChannels(c int) { a.channels = c }

// SetSampleRate sets the sample rate to use when recording from gstreamer. When this method is not called
// the value defaults to 24000.
func (a *Buffer) SetSampleRate(r int) { a.sampleRate = r }

// Start starts the gstreamer processes
func (a *Buffer) Start(codec Codec) error {
	if err := a.setupDevices(); err != nil {
		return err
	}

	a.pbkPipeline = a.buildPlaybackPipeline(codec)
	a.recPipeline = a.buildRecordingPipeline()

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

// Wait will join to the streaming process and block until its finished,
// returning an error if the process exits non-zero.
func (a *Buffer) Wait() error {
	errs := make([]string, 0)
	if err := a.pbkPipeline.Wait(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := a.recPipeline.Wait(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, " : "))
	}
	return nil
}

// Errors returns any errors that ocurred during the streaming process.
func (a *Buffer) Errors() []error {
	errs := make([]error, 0)
	if err := a.pbkPipeline.Error(); err != nil {
		errs = append(errs, err)
	}
	if err := a.recPipeline.Error(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Stderr returns any output from stderr on the streaming process.
func (a *Buffer) Stderr() string {
	return strings.Join([]string{a.pbkPipeline.Stderr(), a.recPipeline.Stderr()}, " : ")
}

// Read implements ReadCloser and returns data from the audio buffer.
func (a *Buffer) Read(p []byte) (int, error) { return a.pbkPipeline.Read(p) }

// Write implements a WriteCloser and writes data to the audio buffer.
func (a *Buffer) Write(p []byte) (int, error) { return a.recPipeline.Write(p) }

// IsClosed returns true if the buffer is closed.
func (a *Buffer) IsClosed() bool {
	return a.pbkPipeline.IsClosed() && a.recPipeline.IsClosed() && a.closed
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
		a.deviceManager.Destroy()
		a.closed = true
	}
	return nil
}
