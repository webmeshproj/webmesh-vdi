// Package audio contains a buffer for streaming audio from a desktop to and from
// a websocket client. It is used by the kvdi-proxy to provide playback and microphone
// support.
package audio

import (
	"io"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

// Buffer provides a ReadWriteCloser for proxying audio data to
// and from a websocket connection. The read-buffer is populated with opus/webm
// data and writes to write-buffer can be in any format that gstreamer `decodebin`
// supports.
type Buffer struct {
	mainLoop                                                       *glib.MainLoop
	logger                                                         logr.Logger
	pbkReader                                                      io.ReadCloser
	recWriter                                                      io.WriteCloser
	micSinkPipeline                                                *gst.Pipeline
	channels, sampleRate, micChannels, micSampleRate               int
	pulseServer, pulseFormat, pulseMonitor, pulseMic, pulseMicPath string
	closed                                                         bool
	wmux                                                           sync.Mutex
	wsize                                                          int
	errChan                                                        chan error
}

// Make sure Buffer implements a io.ReadWriteCloser
var _ io.ReadWriteCloser = &Buffer{}

// NewBuffer returns a new Buffer.
func NewBuffer(opts *BufferOpts) *Buffer {
	gst.Init(nil)
	return &Buffer{
		mainLoop:      glib.NewMainLoop(glib.MainContextDefault(), false),
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
		errChan:       make(chan error),
	}
}

func (a *Buffer) newSinkPipeline() (*gst.Pipeline, error) {
	return newSinkPipeline(
		a.logger.WithName("sink_pipeline"),
		a.errChan,
		&playbackPipelineOpts{
			PulseServer:    a.pulseServer,
			DeviceName:     a.pulseMic,
			SourceFormat:   a.pulseFormat,
			SourceRate:     a.micSampleRate,
			SourceChannels: a.micChannels,
		},
	)
}

func (a *Buffer) newRecordingPipeline() (io.WriteCloser, error) {
	return newRecordingPipelineWriter(
		a.logger.WithName("recording_pipeline"),
		a.errChan,
		&recordingPipelineOpts{
			DeviceFifo:     a.pulseMicPath,
			DeviceFormat:   a.pulseFormat,
			DeviceRate:     a.micSampleRate,
			DeviceChannels: a.micChannels,
		},
	)
}

func (a *Buffer) newPlaybackPipeline() (io.ReadCloser, error) {
	return newPlaybackPipelineReader(
		a.logger.WithName("playback_pipeline"),
		a.errChan,
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

	a.pbkReader, err = a.newPlaybackPipeline()
	if err != nil {
		return err
	}
	a.recWriter, err = a.newRecordingPipeline()
	if err != nil {
		return err
	}

	a.micSinkPipeline, err = a.newSinkPipeline()
	if err != nil {
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
	if err := a.micSinkPipeline.SetState(gst.StatePlaying); err != nil {
		return err
	}

	return nil
}

// RunLoop will run the main loop, blocking until one of the pipelines ends, closes, or any errors.
func (a *Buffer) RunLoop() { a.mainLoop.Run() }

func (a *Buffer) restartRecorder() error {
	a.wmux.Lock()
	defer a.wmux.Unlock()
	var err error
	if err = a.recWriter.Close(); err != nil {
		return err
	}
	a.recWriter, err = a.newRecordingPipeline()
	if err != nil {
		return err
	}
	return nil
}

func (a *Buffer) watchRecPipeline() {
	ticker := time.NewTicker(time.Second * 1)
	lastSize := a.wsize
	lastStartSize := a.wsize
	for range ticker.C {
		// if the playback pipeline is dead, return
		if a.IsClosed() {
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

// Read implements ReadCloser and returns data from the audio buffer.
func (a *Buffer) Read(p []byte) (int, error) {
	select {
	case err := <-a.errChan:
		a.mainLoop.Quit()
		return 0, err
	default:
		n, err := a.pbkReader.Read(p)
		if err != nil {
			a.mainLoop.Quit()
		}
		return n, err
	}
}

// Write implements a WriteCloser and writes data to the audio buffer.
func (a *Buffer) Write(p []byte) (int, error) {
	select {
	case err := <-a.errChan:
		a.mainLoop.Quit()
		return 0, err
	default:
		a.wmux.Lock()
		defer a.wmux.Unlock()
		s, err := a.recWriter.Write(p)
		if err != nil {
			a.mainLoop.Quit()
			return s, err
		}
		a.wsize += s
		return s, nil
	}
}

// IsClosed returns true if the buffer is closed.
func (a *Buffer) IsClosed() bool { return a.closed }

// Close kills the gstreamer pipelines.
func (a *Buffer) Close() error {
	if !a.IsClosed() {
		if err := a.pbkReader.Close(); err != nil {
			return err
		}
		if err := a.recWriter.Close(); err != nil {
			return err
		}
		if err := a.micSinkPipeline.BlockSetState(gst.StateNull); err != nil {
			return err
		}
		a.mainLoop.Quit()
		a.closed = true
	}
	return nil
}
