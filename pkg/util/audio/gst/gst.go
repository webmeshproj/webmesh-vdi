package gst

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

// Pipeline is a helper object for construction gstreamer pipelines.
type Pipeline struct {
	userID string
	pipes  []string
	logger logr.Logger

	cmd    *exec.Cmd
	errOut bytes.Buffer
	writer io.Writer
	reader io.Reader

	ready, closed           bool
	socketSrc               bool
	socketAddr, socketProto string
}

// NewPipeline returns a new empty Pipeline
func NewPipeline(userID string, logger logr.Logger) *Pipeline {
	return &Pipeline{
		pipes:  make([]string, 0),
		userID: userID,
		logger: logger,
	}
}

var pipeSeparator = " ! "

// buildPipeline returns the string representation of the pipeline.
func (g *Pipeline) buildPipeline() string {
	return strings.Join(g.pipes, pipeSeparator)
}

// getGstCmd returns a new command for the GST pipeline.
func (g *Pipeline) getGstCmd() *exec.Cmd {
	fullCmd := fmt.Sprintf("sudo -u \\#%s gst-launch-1.0 -q %s", g.userID, g.buildPipeline())
	g.logger.Info(fmt.Sprintf("Running command: %s", fullCmd))
	return exec.Command("/bin/sh", "-c", fullCmd)
}

// Start starts the GST pipeline
func (g *Pipeline) Start() (err error) {
	g.cmd = g.getGstCmd()

	defer func() { g.ready = true }()

	var errPipe io.ReadCloser

	if g.writer, err = g.cmd.StdinPipe(); err != nil {
		return
	}
	if g.reader, err = g.cmd.StdoutPipe(); err != nil {
		return
	}
	if errPipe, err = g.cmd.StderrPipe(); err != nil {
		return
	}

	go func() {
		if _, err := io.Copy(&g.errOut, errPipe); err != nil {
			g.logger.Error(err, "Erroring reading stderr from gst process")
		}
	}()

	if err = g.cmd.Start(); err != nil {
		return
	}

	return
}

// Pid returns the ID of the GST process.
func (g *Pipeline) Pid() int {
	if g.cmd != nil && g.cmd.Process != nil {
		return g.cmd.Process.Pid
	}
	return 0
}

// Wait waits for the pipeline to finish.
func (g *Pipeline) Wait() error {
	for {
		if g.IsReady() {
			break
		}
	}
	return g.cmd.Wait()
}

// Error returns any errors that occured during the GST pipeline.
func (g *Pipeline) Error() error {
	if g.cmd != nil && g.cmd.ProcessState != nil {
		if g.cmd.ProcessState.Exited() {
			if g.cmd.ProcessState.ExitCode() != 0 {
				return errors.New(g.Stderr())
			}
		}
	}
	return nil
}

// IsClosed returns true if the GST process has exited
func (g *Pipeline) IsClosed() bool {
	if g.cmd == nil || g.cmd.ProcessState == nil {
		return false
	}
	return g.closed && g.cmd.ProcessState.Exited()
}

// Close stops the GST process.
func (g *Pipeline) Close() error {
	if g.cmd == nil {
		return errors.New("Pipeline has not yet started")
	}
	if g.cmd.ProcessState == nil || !g.cmd.ProcessState.Exited() {
		g.logger.Info("Killing pipeline")
		if err := g.cmd.Process.Kill(); err != nil {
			if !strings.HasSuffix(err.Error(), "already finished") {
				return err
			}
		}
	}
	g.closed = true
	return nil
}

// Stderr returns any data on the error output of the GST process.
func (g *Pipeline) Stderr() string {
	return g.errOut.String()
}

// IsReady returns true if the Reader and Writer interfaces are ready to
// send and receive data.
func (g *Pipeline) IsReady() bool {
	return g.ready
}

// Read implements an io.Reader.
func (g *Pipeline) Read(p []byte) (int, error) {
	if g.IsClosed() {
		return 0, fmt.Errorf("Reader interface is closed: %s", g.Stderr())
	}
	for {
		if g.IsReady() {
			break
		}
	}
	return g.reader.Read(p)
}

// Write implements an io.Writer.
func (g *Pipeline) Write(p []byte) (int, error) {
	if g.IsClosed() {
		return 0, fmt.Errorf("Writer interface is closed: %s", g.Stderr())
	}
	for {
		if g.IsReady() {
			break
		}
	}
	return g.writer.Write(p)
}

// WithRawCaps sets the output sample rate, channels, and format for the previous
// raw data in the pipeline.
func (g *Pipeline) WithRawCaps(format string, channels, rate int) *Pipeline {
	return g.WithCaps("audio/x-raw", map[string]interface{}{
		"rate":     rate,
		"channels": channels,
		"format":   strings.ToUpper(format),
	})
}

// WithCaps applies a caps filter to the pipeline.
func (g *Pipeline) WithCaps(mtype string, args map[string]interface{}) *Pipeline {
	if args != nil {
		elems := make([]string, 0)
		for k, v := range args {
			elems = append(elems, fmt.Sprintf("%s=%v", k, v))
		}
		return g.WithPlugin(fmt.Sprintf("%s, %s", mtype, strings.Join(elems, ", ")))
	}
	return g.WithPlugin(mtype)
}

// WithPulseSrc should be called first and sets a pulse server as the source of the
// pipeline.
func (g *Pipeline) WithPulseSrc(userID, deviceName string, channels, sampleRate int) *Pipeline {
	return g.
		WithPlugin(fmt.Sprintf("pulsesrc server=/run/user/%s/pulse/native device=%s", userID, deviceName)).
		WithRawCaps("s16le", channels, sampleRate)
}

// WithFdSrc should be called first and sets an open file descriptor as the
// source of the pipeline. This will usually be used with fd0 to allow using the Pipeline
// as an io.Writer.
func (g *Pipeline) WithFdSrc(fd int, doTimestamp bool) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("fdsrc fd=%d do-timestamp=%v", fd, doTimestamp))
}

// WithTCPSrc should be called first and creates a TCP listener as the
// source of the pipeline.
func (g *Pipeline) WithTCPSrc(host string, port int, doTimestamp bool) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("tcpserversrc host=%s port=%d do-timestamp=%v", host, port, doTimestamp))
}

// WithUDPSrc should be called first and creates a UDP listener as the
// source of the pipeline.
func (g *Pipeline) WithUDPSrc(host string, port int) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("udpsrc address=%s port=%d", host, port))
}

// WithVorbisEncode encodes the next step in the pipeline to vorbis.
func (g *Pipeline) WithVorbisEncode() *Pipeline {
	return g.WithPlugin("vorbisenc")
}

// WithVorbisDecode decodes vorbis data from the previous step in the pipeline.
func (g *Pipeline) WithVorbisDecode() *Pipeline {
	return g.WithPlugin("vorbisdec")
}

// WithOpusEncode encodes the next step in the pipeline to opus.
func (g *Pipeline) WithOpusEncode() *Pipeline {
	return g.WithPlugin("opusenc")
}

// WithOpusDecode decodes opus data from the previous step in the pipeline.
func (g *Pipeline) WithOpusDecode(plc bool) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("opusdec plc=%v", plc))
}

// WithLameEncode encodes the next step in the pipeline to mp3.
func (g *Pipeline) WithLameEncode() *Pipeline {
	return g.WithPlugin("lamemp3enc")
}

// WithLameDecode decodes mp3 data from the previous step in the pipeline.
func (g *Pipeline) WithLameDecode() *Pipeline {
	g = g.WithPlugin("mpegaudioparse")
	return g.WithPlugin("mpg123audiodec")
}

// WithWavEncode encodes the next step in the pipeline to wav.
func (g *Pipeline) WithWavEncode() *Pipeline {
	return g.WithPlugin("wavenc")
}

// WithOggMux wraps the next step in the pipeline in an ogg container.
func (g *Pipeline) WithOggMux() *Pipeline {
	return g.WithPlugin("oggmux")
}

// WithOggDemux unwraps an ogg container from the previous step in the pipeline.
func (g *Pipeline) WithOggDemux() *Pipeline {
	return g.WithPlugin("oggdemux")
}

// WithWebmMux wraps the next step in the pipeline in a webm container.
func (g *Pipeline) WithWebmMux() *Pipeline {
	return g.WithPlugin("webmmux")
}

// WithCutter tries to remove radio silence from the previous step in the pipeline.
func (g *Pipeline) WithCutter() *Pipeline {
	return g.WithPlugin("cutter")
}

// WithAudioConvert tries to automatically decode the data from the previous step
// in the pipeline.
func (g *Pipeline) WithAudioConvert() *Pipeline {
	return g.WithPlugin("audioconvert")
}

// WithAudioResample resamples the audio data from the previous step in the pipeline.
func (g *Pipeline) WithAudioResample() *Pipeline {
	return g.WithPlugin("audioresample")
}

// WithFdSink is called last and writes the data from the pipeline to an
// open file descriptor.
func (g *Pipeline) WithFdSink(fd int) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("fdsink fd=%d", fd))
}

// WithFileSink is called last and writes the data from the pipeline to
// the given file.
func (g *Pipeline) WithFileSink(file string, appendFile bool) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("filesink append=%v location=%s", appendFile, file))
}

// WithUDPSink is called last and writes the data from the pipeline to a
// UDP socket.
func (g *Pipeline) WithUDPSink(host string, port int) *Pipeline {
	return g.WithPlugin(fmt.Sprintf("udpsink host=%s port=%d", host, port))
}

// WithPlugin adds the given plugin (and args) to the pipeline
func (g *Pipeline) WithPlugin(pipe string) *Pipeline {
	g.pipes = append(g.pipes, pipe)
	return g
}
