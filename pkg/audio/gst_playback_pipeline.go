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
	pulsesrc = "pulsesrc"
	fdsink   = "fdsink"
	cutter   = "cutter"
	opusenc  = "opusenc"
	webmmux  = "webmmux"
)

func (p *PlaybackPipeline) setupPipeline() error {
	// Build all the elements
	pulseCaps := gst.NewRawCaps(p.opts.SourceFormat, p.opts.SourceRate, p.opts.SourceChannels)
	encoderElements, err := p.NewElementMany(pulsesrc, cutter, opusenc, webmmux, fdsink)
	if err != nil {
		return err
	}
	if err := encoderElements[fdsink].Set("fd", int(p.WriterFd())); err != nil {
		return err
	}
	if err := encoderElements[pulsesrc].Set("server", p.opts.PulseServer); err != nil {
		return err
	}
	if err := encoderElements[pulsesrc].Set("device", p.opts.DeviceName); err != nil {
		return err
	}

	// Add all the elements to the pipeline
	if err := p.BinAddMany(
		encoderElements[pulsesrc],
		encoderElements[cutter],
		encoderElements[opusenc],
		encoderElements[webmmux],
		encoderElements[fdsink],
	); err != nil {
		return err
	}

	// Link the pulsesrc to cutter with caps
	if err := p.ElementLinkFiltered(encoderElements[pulsesrc], encoderElements[cutter], pulseCaps); err != nil {
		return err
	}

	// Link the rest of the pipeline
	if err := p.ElementLinkMany(
		encoderElements[cutter],
		encoderElements[opusenc],
		encoderElements[webmmux],
		encoderElements[fdsink],
	); err != nil {
		return err
	}

	return nil
}
