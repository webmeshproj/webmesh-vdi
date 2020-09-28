// Package audio contains a buffer for streaming audio from a desktop to and from
// a websocket client. It is used by the kvdi-proxy to provide playback and microphone
// support.
package audio

import (
	"io"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/gstauto"
)

// Buffer provides a ReadWriteCloser for proxying audio data to
// and from a websocket connection. The read-buffer is populated with opus/webm
// data and writes to write-buffer can be in any format that gstreamer `decodebin`
// supports.
type Buffer struct {
	logger                                                         logr.Logger
	pbkPipeline                                                    *gstauto.PipelineReaderSimple
	recPipeline                                                    *gstauto.PipelineWriterSimple
	micSinkPipeline                                                *gstauto.PipelinerSimple
	channels, sampleRate, micChannels, micSampleRate               int
	pulseServer, pulseFormat, pulseMonitor, pulseMic, pulseMicPath string
	closed                                                         bool
	wmux                                                           sync.Mutex
	wsize                                                          int
}

// Make sure Buffer implements a io.ReadWriteCloser
var _ io.ReadWriteCloser = &Buffer{}

// NewBuffer returns a new Buffer.
func NewBuffer(opts *BufferOpts) *Buffer {
	gst.Init(nil)
	return &Buffer{
		logger:        opts.getLogger(),
		pulseServer:   opts.getPulseServer(),
		pulseFormat:   opts.getPulseFormat(),
		channels:      opts.getPulseMonitorChannels(),
		micChannels:   opts.getPulseMicChannels(),
		micSampleRate: opts.getPulseMicRate(),
		sampleRate:    opts.getPulsePlaybackRate(),
		pulseMonitor:  opts.getPulseMonitorName(),
		pulseMic:      opts.getMicName(),
		pulseMicPath:  opts.getMicPath(),
	}
}

func (a *Buffer) newSinkPipeline() (*gstauto.PipelinerSimple, error) {
	return newSinkPipeline(
		a.logger.WithName("mic_null_monitor"),
		&playbackPipelineOpts{
			PulseServer:    a.pulseServer,
			DeviceName:     a.pulseMic,
			SourceFormat:   a.pulseFormat,
			SourceRate:     a.micSampleRate,
			SourceChannels: a.micChannels,
		},
	)
}

func (a *Buffer) newRecordingPipeline() (*gstauto.PipelineWriterSimple, error) {
	return newRecordingPipeline(
		a.logger.WithName("gst_recorder"),
		&recordingPipelineOpts{
			DeviceFifo:     a.pulseMicPath,
			DeviceFormat:   a.pulseFormat,
			DeviceRate:     a.micSampleRate,
			DeviceChannels: a.micChannels,
		},
	)
}

func (a *Buffer) newPlaybackPipeline() (*gstauto.PipelineReaderSimple, error) {
	return newPlaybackPipeline(
		a.logger.WithName("gst_playback"),
		&playbackPipelineOpts{
			PulseServer:    a.pulseServer,
			DeviceName:     a.pulseMonitor,
			SourceFormat:   a.pulseFormat,
			SourceRate:     a.sampleRate,
			SourceChannels: a.channels,
		},
	)
}

// Start starts the gstreamer processes
func (a *Buffer) Start() error {
	var err error

	a.pbkPipeline, err = a.newPlaybackPipeline()
	if err != nil {
		return err
	}
	a.recPipeline, err = a.newRecordingPipeline()
	if err != nil {
		return err
	}

	a.micSinkPipeline, err = a.newSinkPipeline()
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

	// Watch the mic pipeline and restart it if there is ever more than a
	// one second period of silence. This is a workaround to inferring that
	// the user has toggled audio. There should really be a way to signal that
	// to this process instead, since a trigger happy user can bash the toggle
	// enough to ultimately cause a race.
	go a.watchRecPipeline()

	// Start a dump of the contents on the mic device.
	// This allows PulseAudio to flush the buffer on the device so when other
	// applications request audio from it they don't get dumped the entire history
	// at once.
	if err := a.micSinkPipeline.Start(); err != nil {
		return err
	}

	return nil
}

func (a *Buffer) restartRecorder() error {
	a.wmux.Lock()
	defer a.wmux.Unlock()
	var err error
	if err = a.recPipeline.Close(); err != nil {
		return err
	}
	a.recPipeline, err = a.newRecordingPipeline()
	if err != nil {
		return err
	}
	return a.recPipeline.Start()
}

func (a *Buffer) watchRecPipeline() {
	ticker := time.NewTicker(time.Second * 1)
	lastSize := a.wsize
	lastStartSize := a.wsize
	for range ticker.C {
		// if the playback pipeline is dead, return
		if a.pbkPipeline.Pipeline().GetState() == gst.StateNull {
			return
		}
		if a.wsize == lastSize {
			if lastStartSize == a.wsize {
				// we have restarted already and there is no data still yet
				continue
			}
			a.logger.Info("Restarting recording pipeline")
			if err := a.restartRecorder(); err != nil {
				a.logger.Error(err, "Failed to restart recording pipeline")
				return
			}
			a.logger.Info("Recording pipeline restarted")
			lastStartSize = a.wsize
		}
		lastSize = a.wsize
	}
}

// Wait will wait for the main playback pipeline to complete.
func (a *Buffer) Wait() {
	gst.Wait(a.pbkPipeline.Pipeline())
}

// Read implements ReadCloser and returns data from the audio buffer.
func (a *Buffer) Read(p []byte) (int, error) { return a.pbkPipeline.Read(p) }

// Write implements a WriteCloser and writes data to the audio buffer.
func (a *Buffer) Write(p []byte) (int, error) {
	a.wmux.Lock()
	defer a.wmux.Unlock()
	s, err := a.recPipeline.Write(p)
	if err != nil {
		return s, err
	}
	a.wsize += s
	return s, nil
}

// IsClosed returns true if the buffer is closed.
func (a *Buffer) IsClosed() bool {
	return a.closed
}

// Close kills the gstreamer pipelines.
func (a *Buffer) Close() error {
	if !a.IsClosed() {
		if err := a.pbkPipeline.Close(); err != nil {
			return err
		}
		if err := a.recPipeline.Close(); err != nil {
			return err
		}
		if err := a.micSinkPipeline.Pipeline().Destroy(); err != nil {
			return err
		}
		a.closed = true
	}
	return nil
}
