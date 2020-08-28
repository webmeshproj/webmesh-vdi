package audio

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var testLogger = logf.Log.WithName("test")

func TestNewBuffer(t *testing.T) {
	buffer := NewBuffer(testLogger, "9000")
	if buffer.userID != "9000" {
		t.Error("Expected buffer user id to be 9000, got:", buffer.userID)
	}
}

func TestBuildGSTPipeline(t *testing.T) {
	buffer := NewBuffer(testLogger, "9000")

	prefix := fmt.Sprintf("sudo -u audioproxy gst-launch-1.0 -q pulsesrc server=/run/user/%s/pulse/native ! audio/x-raw, channels=2, rate=24000", "9000")
	suffix := "fdsink fd=1"

	testCases := []struct {
		codec    Codec
		expected string
	}{
		{
			codec:    CodecVorbis,
			expected: fmt.Sprintf("%s ! vorbisenc ! oggmux ! %s", prefix, suffix),
		},
		{
			codec:    CodecOpus,
			expected: fmt.Sprintf("%s ! cutter ! opusenc ! webmmux ! %s", prefix, suffix),
		},
		{
			codec:    CodecMP3,
			expected: fmt.Sprintf("%s ! lamemp3enc ! %s", prefix, suffix),
		},
		{
			// default to opus
			codec:    Codec("Unknown"),
			expected: fmt.Sprintf("%s ! cutter ! opusenc ! webmmux ! %s", prefix, suffix),
		},
	}

	for _, tc := range testCases {
		if pipeline := buffer.buildPipeline(tc.codec); pipeline != tc.expected {
			t.Errorf("Expected %s for %s, got: %s", tc.expected, tc.codec, pipeline)
		}
	}
}

func TestUseBuffer(t *testing.T) {
	buffer := NewBuffer(testLogger, "9000")

	// override exec method to run controlled commands, depends on tests
	// executed in an environment with a shell.
	// Write the word "testing" to stdout and stderr.
	buffer.exec = func(string, ...string) *exec.Cmd {
		return exec.Command("/bin/sh", "-c", "echo -n testing | tee /dev/stderr")
	}

	if err := buffer.Start(CodecOpus); err != nil {
		t.Error("Expected no error starting buffer, got:", err.Error())
	}

	stdout, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal(err)
	}
	if string(stdout) != "testing" {
		t.Error("Expected buffer to have word 'testing', got:", string(stdout))
	}

	stderr := buffer.Stderr()
	if stderr != "testing" {
		t.Error("Expected stderr to include word 'testing', got:", string(stderr))
	}

	if err := buffer.Wait(); err != nil {
		t.Error("Expected no error waiting for process to finish, got:", err.Error())
	}

	if !buffer.IsClosed() {
		t.Error("Expected buffer to be closed, got false")
	}

	if err := buffer.Close(); err != nil {
		t.Error("Expected no error closing the buffer anyway, got:", err.Error())
	}

	if err := buffer.Error(); err != nil {
		t.Error("Expected no error on buffer, got:", err.Error())
	}
}
