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

// RecordingPipeline implements a WriteCloser that writes raw audio data to
// a virtual mic on a pulse server. It is assumed the audio source is parseable
// by gst decodebin.
type RecordingPipeline struct {
	*Pipeline
}

// NewRecordingPipeline returns a new RecordingPipeline. For now the pipeline is construced using
// `gst_parse_launch`. However, this should be refactored to gain more control over latency.
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
