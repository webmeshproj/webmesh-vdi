package audio

import (
	"github.com/go-logr/logr"
	"github.com/tinyzimmer/kvdi/pkg/audio/gst"
)

// PlaybackPipelineOpts are options passed to the playback pipeline.
type PlaybackPipelineOpts struct {
	PulseServer, DeviceName, SourceFormat string
	SourceRate, SourceChannels            int
}

// PlaybackPipeline implements a ReadCloser that reads raw audio data from a
// pulseaudio server, encodes it to opus/webm, and makes it available on an
// internal buffer for reading.
type PlaybackPipeline struct {
	*gst.Pipeline
	opts *PlaybackPipelineOpts
}

// NewPlaybackPipeline returns a new PlaybackPipeline.
func NewPlaybackPipeline(logger logr.Logger, opts *PlaybackPipelineOpts) (*PlaybackPipeline, error) {
	pipeline, err := gst.NewPipeline(logger)
	if err != nil {
		return nil, err
	}
	pbkPipeline := &PlaybackPipeline{
		Pipeline: pipeline,
		opts:     opts,
	}
	if err := pbkPipeline.setupPipeline(); err != nil {
		return nil, err
	}
	return pbkPipeline, nil
}

const (
	cutter  = "cutter"
	opusenc = "opusenc"
	webmmux = "webmmux"
)

func (p *PlaybackPipeline) setupPipeline() error {
	// Build all the elements
	pulseSrc, err := gst.NewPulseSrc(p.opts.PulseServer, p.opts.DeviceName)
	if err != nil {
		return err
	}
	pulseCaps, err := gst.NewPulseCaps(p.opts.SourceFormat, p.opts.SourceRate, p.opts.SourceChannels)
	if err != nil {
		return err
	}
	encoderElements, err := p.NewElementMany(cutter, opusenc, webmmux)
	if err != nil {
		return err
	}
	fdSink, err := gst.NewFdSink(int(p.WriterFd()))
	if err != nil {
		return err
	}

	// Add all the elements to the pipeline
	if err := p.BinAddMany(
		pulseSrc,
		encoderElements[cutter],
		encoderElements[opusenc],
		encoderElements[webmmux],
		fdSink,
	); err != nil {
		return err
	}

	// Link the pulsesrc to cutter with caps
	if err := p.ElementLinkFiltered(pulseSrc, encoderElements[cutter], pulseCaps); err != nil {
		return err
	}

	// Link the rest of the pipeline
	if err := p.ElementLinkMany(
		encoderElements[cutter],
		encoderElements[opusenc],
		encoderElements[webmmux],
		fdSink,
	); err != nil {
		return err
	}

	return nil
}
