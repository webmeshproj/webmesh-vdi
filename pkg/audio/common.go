package audio

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

func newRawCaps(format string, rate, channels int) *gst.Caps {
	return gst.NewCapsFromString(
		fmt.Sprintf("audio/x-raw,format=%s,rate=%d,channels=%d", format, rate, channels),
	)
}
