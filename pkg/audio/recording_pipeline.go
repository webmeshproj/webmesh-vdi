package audio

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/gstauto"
)

// RecordingPipelineOpts are options passed to the playback pipeline.
type recordingPipelineOpts struct {
	DeviceFifo, DeviceFormat   string
	DeviceRate, DeviceChannels int
}

// NewRecordingPipeline returns a new RecordingPipeline. For now the pipeline is construced using
// `gst_parse_launch`. However, this should be refactored to gain more control over latency.
func newRecordingPipeline(logger logr.Logger, opts *recordingPipelineOpts) (*gstauto.PipelineWriterSimple, error) {
	// TODO: decodebin requires dynamic linking so a little more complex than playback
	// Though would like more control over pads in this pipeline to try to reduce latency
	return gstauto.NewPipelineWriterSimpleFromString(newPipelineStringFromOpts(opts))
}

func newPipelineStringFromOpts(opts *recordingPipelineOpts) string {
	return fmt.Sprintf(
		"decodebin ! audioconvert ! audioresample ! audio/x-raw, format=%s, rate=%d, channels=%d ! filesink location=%s append=true",
		opts.DeviceFormat,
		opts.DeviceRate,
		opts.DeviceChannels,
		opts.DeviceFifo,
	)
}

// NewSinkPipeline returns a pipeline that dumps audio data to a null device as fast as possible.
// This is useful for flushing the contents of a mic buffer when there are no applications listening
// to it.
func newSinkPipeline(logger logr.Logger, opts *playbackPipelineOpts) (*gstauto.PipelinerSimple, error) {
	cfg := &gstauto.PipelineConfig{
		Elements: []*gstauto.PipelineElement{
			{
				Name:     "pulsesrc",
				Data:     map[string]interface{}{"server": opts.PulseServer, "device": opts.DeviceName},
				SinkCaps: gst.NewRawCaps(opts.SourceFormat, opts.SourceRate, opts.SourceChannels),
			},
			{
				Name: "fakesink",
				Data: map[string]interface{}{"sync": false},
			},
		},
	}
	return gstauto.NewPipelinerSimpleFromConfig(cfg)
}
