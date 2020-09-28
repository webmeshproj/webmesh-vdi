package audio

import (
	"github.com/go-logr/logr"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/gstauto"
)

// PlaybackPipelineOpts are options passed to the playback pipeline.
type playbackPipelineOpts struct {
	PulseServer, DeviceName, SourceFormat string
	SourceRate, SourceChannels            int
}

// NewPlaybackPipeline returns a new Pipeline for audio playback.
func newPlaybackPipeline(logger logr.Logger, opts *playbackPipelineOpts) (*gstauto.PipelineReaderSimple, error) {
	cfg := &gstauto.PipelineConfig{
		Elements: []*gstauto.PipelineElement{
			{
				Name: "pulsesrc",
				Data: map[string]interface{}{
					"server": opts.PulseServer,
					"device": opts.DeviceName,
				},
				SinkCaps: gst.NewRawCaps(opts.SourceFormat, opts.SourceRate, opts.SourceChannels),
			},
			{Name: "cutter"},
			{Name: "opusenc"},
			{Name: "webmmux"},
		},
	}
	return gstauto.NewPipelineReaderSimpleFromConfig(cfg)
}
