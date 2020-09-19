package gst

import (
	"fmt"

	"github.com/go-logr/logr"
)

// RecordingPipelineOpts are options passed to the playback pipeline.
type RecordingPipelineOpts struct {
	DeviceFifo, DeviceFormat   string
	DeviceRate, DeviceChannels int
}

// RecordingPipeline implements a ReadCloser that reads raw audio data from a
// pulseaudio server, encodes it to opus/webm, and makes it available on an
// internal buffer for reading.
type RecordingPipeline struct {
	*Pipeline
}

// NewRecordingPipeline returns a new RecordingPipeline.
func NewRecordingPipeline(logger logr.Logger, opts *RecordingPipelineOpts) (*RecordingPipeline, error) {
	// TODO: decodebin required dynamic linking so a little more complex than playback
	// Though would like more control over pads in this pipeline to try to reduce latency
	pipelineString := fmt.Sprintf(
		"decodebin ! audioconvert ! audioresample ! audio/x-raw, format=%s, rate=%d, channels=%d ! filesink location=%s append=true",
		opts.DeviceFormat,
		opts.DeviceRate,
		opts.DeviceChannels,
		opts.DeviceFifo,
	)
	pipeline, err := NewPipelineFromLaunchString(logger, pipelineString, true, false)
	if err != nil {
		return nil, err
	}
	recPipeline := &RecordingPipeline{pipeline}
	return recPipeline, nil
}
