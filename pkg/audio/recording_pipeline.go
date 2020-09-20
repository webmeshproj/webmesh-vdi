package audio

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/tinyzimmer/kvdi/pkg/audio/gst"
)

// RecordingPipelineOpts are options passed to the playback pipeline.
type RecordingPipelineOpts struct {
	DeviceFifo, DeviceFormat   string
	DeviceRate, DeviceChannels int
}

// NewRecordingPipeline returns a new RecordingPipeline. For now the pipeline is construced using
// `gst_parse_launch`. However, this should be refactored to gain more control over latency.
func NewRecordingPipeline(logger logr.Logger, opts *RecordingPipelineOpts) (*gst.Pipeline, error) {
	// TODO: decodebin required dynamic linking so a little more complex than playback
	// Though would like more control over pads in this pipeline to try to reduce latency
	return gst.NewPipelineFromLaunchString(logger, newPipelineStringFromOpts(opts), true, false)
}

func newPipelineStringFromOpts(opts *RecordingPipelineOpts) string {
	return fmt.Sprintf(
		"decodebin ! audioconvert ! audioresample ! audio/x-raw, format=%s, rate=%d, channels=%d ! filesink location=%s append=true",
		opts.DeviceFormat,
		opts.DeviceRate,
		opts.DeviceChannels,
		opts.DeviceFifo,
	)
}
