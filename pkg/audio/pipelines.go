//go:build audio

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

package audio

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

// PlaybackPipelineOpts are options passed to the playback pipeline.
type playbackPipelineOpts struct {
	PulseServer, DeviceName, SourceFormat string
	SourceRate, SourceChannels            int
}

type pipelineReader struct {
	rPipe    *io.PipeReader
	wPipe    *io.PipeWriter
	pipeline *gst.Pipeline
}

func newPlaybackPipelineReader(log logr.Logger, errors chan error, opts *playbackPipelineOpts) (rdr io.ReadCloser, err error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return
	}

	elements, err := gst.NewElementMany("pulsesrc", "cutter", "opusenc", "webmmux", "appsink")
	if err != nil {
		return
	}
	pulsesrc, cutter, opusenc, webmmux, appsink := elements[0], elements[1], elements[2], elements[3], elements[4]

	if err = pulsesrc.SetProperty("server", opts.PulseServer); err != nil {
		return
	}

	deviceName := opts.DeviceName
	if !strings.HasSuffix(deviceName, ".monitor") {
		deviceName = fmt.Sprintf("%s.monitor", deviceName)
	}

	if err = pulsesrc.SetProperty("device", deviceName); err != nil {
		return
	}

	pulsecaps := newRawCaps(opts.SourceFormat, opts.SourceRate, opts.SourceChannels)

	r, w := io.Pipe()

	app.SinkFromElement(appsink).SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(self *app.Sink) gst.FlowReturn {
			sample := self.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}
			if _, err := io.Copy(w, buffer.Reader()); err != nil {
				return gst.FlowError
			}
			return gst.FlowOK
		},
	})

	if err = pipeline.AddMany(elements...); err != nil {
		return
	}
	if err = pulsesrc.LinkFiltered(cutter, pulsecaps); err != nil {
		return
	}
	if err = gst.ElementLinkMany(cutter, opusenc, webmmux, appsink); err != nil {
		return
	}

	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			merr := msg.ParseError()
			log.Error(merr, "Error from pipeline", "Debug", merr.DebugString())
			errors <- merr
		case gst.MessageEOS:
			log.Info("Pipeline has reached EOS")
			errors <- app.ErrEOS
		case gst.MessageElement:
		default:
			log.Info(msg.String())
		}
		return true
	})

	if err = pipeline.SetState(gst.StatePlaying); err != nil {
		return
	}

	return &pipelineReader{
		rPipe:    r,
		wPipe:    w,
		pipeline: pipeline,
	}, nil
}

func (r *pipelineReader) Read(p []byte) (int, error) {
	return r.rPipe.Read(p)
}

func (r *pipelineReader) Close() error {
	if err := r.pipeline.BlockSetState(gst.StateNull); err != nil {
		return err
	}
	return r.wPipe.Close()
}

// RecordingPipelineOpts are options passed to the playback pipeline.
type recordingPipelineOpts struct {
	DeviceFifo, DeviceFormat   string
	DeviceRate, DeviceChannels int
}

type pipelineWriter struct {
	wReader, wWriter *os.File
	wBuf             *bufio.Writer
	pipeline         *gst.Pipeline
}

func newRecordingPipelineWriter(log logr.Logger, errors chan error, opts *recordingPipelineOpts) (wrtr io.WriteCloser, err error) {
	r, w, err := os.Pipe()
	if err != nil {
		return
	}
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return
	}
	elements, err := gst.NewElementMany("fdsrc", "decodebin")
	if err != nil {
		return
	}

	if err = elements[0].SetProperty("fd", int(r.Fd())); err != nil {
		return
	}

	_, err = elements[1].Connect("pad-added", func(self *gst.Element, srcPad *gst.Pad) {
		newElements, err := gst.NewElementMany("queue", "audioconvert", "audioresample", "filesink")
		if err != nil {
			self.ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed, err.Error(), "")
			return
		}
		queue, audioconvert, audioresample, filesink := newElements[0], newElements[1], newElements[2], newElements[3]

		resampleCaps := newRawCaps(opts.DeviceFormat, opts.DeviceRate, opts.DeviceChannels)
		if err := filesink.SetProperty("location", opts.DeviceFifo); err != nil {
			self.ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed, err.Error(), "")
		}
		if err := filesink.SetProperty("append", true); err != nil {
			self.ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed, err.Error(), "")
		}

		if err := pipeline.AddMany(queue, audioconvert, audioresample, filesink); err != nil {
			self.ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed, err.Error(), "")
		}
		if err := gst.ElementLinkMany(queue, audioconvert, audioresample); err != nil {
			self.ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed, err.Error(), "")
		}
		if err := audioresample.LinkFiltered(filesink, resampleCaps); err != nil {
			self.ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed, err.Error(), "")
		}

		for _, e := range newElements {
			e.SyncStateWithParent()
		}

		srcPad.Link(queue.GetStaticPad("sink"))
	})
	if err != nil {
		return
	}

	if err = pipeline.AddMany(elements...); err != nil {
		return
	}
	if err = gst.ElementLinkMany(elements...); err != nil {
		return
	}

	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			merr := msg.ParseError()
			log.Error(merr, "Error from pipeline", "Debug", merr.DebugString())
			errors <- merr
		case gst.MessageEOS:
			log.Info("Pipeline has reached EOS")
			errors <- app.ErrEOS
		case gst.MessageElement:
		default:
			log.Info(msg.String())
		}
		return true
	})

	if err = pipeline.SetState(gst.StatePlaying); err != nil {
		return
	}

	return &pipelineWriter{
		wReader:  r,
		wWriter:  w,
		wBuf:     bufio.NewWriter(w),
		pipeline: pipeline,
	}, nil
}

func (w *pipelineWriter) Write(p []byte) (int, error) { return w.wBuf.Write(p) }

func (w *pipelineWriter) Close() error {
	if err := w.wWriter.Close(); err != nil {
		return err
	}
	if err := w.pipeline.BlockSetState(gst.StateNull); err != nil {
		return err
	}
	return w.wReader.Close()
}

// NewSinkPipeline returns a pipeline that dumps audio data to a null device as fast as possible.
// This is useful for flushing the contents of a mic buffer when there are no applications listening
// to it.
func newSinkPipeline(log logr.Logger, errors chan error, opts *playbackPipelineOpts) (pipeline *gst.Pipeline, err error) {
	pipeline, err = gst.NewPipeline("")
	if err != nil {
		return
	}
	elements, err := gst.NewElementMany("pulsesrc", "fakesink")
	if err != nil {
		return
	}

	pulsesrc, fakesink := elements[0], elements[1]
	pulsecaps := newRawCaps(opts.SourceFormat, opts.SourceRate, opts.SourceChannels)

	if err = pulsesrc.SetProperty("server", opts.PulseServer); err != nil {
		return
	}
	if err = pulsesrc.SetProperty("device", opts.DeviceName); err != nil {
		return
	}
	if err = fakesink.SetProperty("sync", false); err != nil {
		return
	}

	if err = pipeline.AddMany(pulsesrc, fakesink); err != nil {
		return
	}
	if err = pulsesrc.LinkFiltered(fakesink, pulsecaps); err != nil {
		return
	}

	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			merr := msg.ParseError()
			log.Error(merr, "Error from pipeline", "Debug", merr.DebugString())
			errors <- merr
		case gst.MessageEOS:
			log.Info("Pipeline has reached EOS")
			errors <- app.ErrEOS
		case gst.MessageElement:
		default:
			log.Info(msg.String())
		}
		return true
	})

	return pipeline, nil
}

func newRawCaps(format string, rate, channels int) *gst.Caps {
	return gst.NewCapsFromString(
		fmt.Sprintf("audio/x-raw,format=%s,rate=%d,channels=%d", format, rate, channels),
	)
}
