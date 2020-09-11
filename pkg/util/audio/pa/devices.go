// Package pa implements a wrapper around using pactl to setup virtual audio devices.
//
// The intention is to eventually convert this to using native PA APIs.
package pa

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/go-logr/logr"
)

// DeviceManager is an object for managing virtual PulseAudio devices.
type DeviceManager struct {
	userID    string
	deviceIDs []string
	logger    logr.Logger

	mux sync.Mutex
}

// NewDeviceManager returns a new DeviceManager.
func NewDeviceManager(logger logr.Logger, userID string) *DeviceManager {
	return &DeviceManager{
		userID:    userID,
		deviceIDs: make([]string, 0),
		logger:    logger,
	}
}

func (p *DeviceManager) runPaCtlCmd(cmd string) ([]byte, error) {
	fullCmd := fmt.Sprintf("sudo -u \\#%s pactl -s /run/user/%s/pulse/native %s", p.userID, p.userID, cmd)
	fullCmd = strings.Replace(strings.Replace(fullCmd, "\\\n", "", -1), "\t", "", -1)
	p.logger.Info(fmt.Sprintf("Running command: %s", fullCmd))
	return exec.Command("/bin/sh", "-c", fullCmd).CombinedOutput()
}

func (p *DeviceManager) appendDevice(stdoutID []byte) {
	p.deviceIDs = append(p.deviceIDs, strings.TrimSpace(string(stdoutID)))
}

func (p *DeviceManager) addDeviceCmd(cmd string) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	out, err := p.runPaCtlCmd(cmd)
	if err != nil {
		return err
	}
	p.appendDevice(out)
	return nil
}

// AddSink adds a new null-sink with the given name and description.
func (p *DeviceManager) AddSink(name, description string) error {
	cmd := fmt.Sprintf(
		`load-module module-null-sink sink_name=%s sink_properties=device.description="%s"`,
		name, description,
	)
	return p.addDeviceCmd(cmd)
}

// AddSource adds a new pipe-source with the given name, description, and FIFO path.
func (p *DeviceManager) AddSource(name, description, file, format string, channels, rate int) error {
	cmd := fmt.Sprintf(`load-module module-pipe-source \
	source_name=%s \
	source_properties=device.description="%s" \
	file=%s \
	format=%s rate=%d channels=%d`,
		name, description, file, format, rate, channels,
	)
	return p.addDeviceCmd(cmd)
}

// SetDefaultSource will set the default source for recording clients.
func (p *DeviceManager) SetDefaultSource(name string) error {
	cmd := fmt.Sprintf("set-default-source %s", name)
	_, err := p.runPaCtlCmd(cmd)
	return err
}

// Destroy will delete all currently managed PA devices. Replace the device IDs list
// so that Destroy is idempotent.
func (p *DeviceManager) Destroy() {
	p.mux.Lock()
	defer p.mux.Unlock()
	newDeviceIDs := make([]string, 0)
	for _, dev := range p.deviceIDs {
		if _, err := p.runPaCtlCmd(fmt.Sprintf("unload-module %s", dev)); err != nil {
			p.logger.Error(err, fmt.Sprintf("Error removing device %s", dev))
			newDeviceIDs = append(newDeviceIDs, dev)
			continue
		}
	}
	p.deviceIDs = newDeviceIDs
}
