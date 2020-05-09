package audio

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type AudioBuffer struct {
	cmd    *exec.Cmd
	buffer io.ReadCloser
	stderr bytes.Buffer
	closed bool
}

var _ io.ReadCloser = &AudioBuffer{}

func NewBuffer() *AudioBuffer {
	return &AudioBuffer{}
}

func (a *AudioBuffer) Start(outFmt string) error {
	switch outFmt {
	case "ogg":
		a.cmd = exec.Command("/bin/sh", "-c", "sudo -u audioproxy parec -s /run/user/9000/pulse/native | oggenc -b 192 -o - --raw -")
	case "mp3":
		a.cmd = exec.Command("/bin/sh", "-c", "sudo -u audioproxy parec -s /run/user/9000/pulse/native --passthrough | lame -r -V0 -")
	default:
		a.cmd = exec.Command("/bin/sh", "-c", "sudo -u audioproxy parec -s /run/user/9000/pulse/native --passthrough | lame -r -V0 -")
	}

	var err error
	a.buffer, err = a.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	errPipe, err := a.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		if _, err := io.Copy(&a.stderr, errPipe); err != nil {
			fmt.Println("Error reading stderr from recorder proceess:", err)
		}
	}()

	if err := a.cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := a.cmd.Wait(); err != nil {
			fmt.Println("Process exited with error:", err)
		}
		a.closed = true
	}()

	return nil
}

func (a *AudioBuffer) Stderr() string { return a.stderr.String() }

func (a *AudioBuffer) Read(p []byte) (int, error) {
	return a.buffer.Read(p)
}

func (a AudioBuffer) IsClosed() bool { return a.closed }

func (a *AudioBuffer) Close() error {
	if !a.IsClosed() {
		return a.cmd.Process.Kill()
	}
	return nil
}
