/*

   Copyright 2020,2021 Avi Zimmerman

   This file is part of kvdi.

   kvdi is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   kvdi is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with kvdi.  If not, see <https://www.gnu.org/licenses/>.

*/

package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/kvdi/kvdi/pkg/audio/pa"
)

type pulsefifosink struct {
	settings *settings
	state    *state
}

var capsTemplate = gst.NewCapsFromString("audio/x-raw, format=(string){ S16LE, S16BE, S32LE, S32BE, S24LE, S24BE }, layout=(string)interleaved, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ]")

var staticSinkPadTemplate = gst.NewPadTemplate(
	"sink",
	gst.PadDirectionSink,
	gst.PadPresenceAlways,
	capsTemplate,
)

func (p *pulsefifosink) New() glib.GoObjectSubclass {
	cat.LogInfo("Initializing new pulsefifosink object")
	return &pulsefifosink{
		settings: defaultSettings(),
		state:    &state{},
	}
}

func (p *pulsefifosink) ClassInit(klass *glib.ObjectClass) {
	cat.LogInfo("Initializing gofilesink class")
	class := gst.ToElementClass(klass)
	class.SetMetadata(
		"Pulse FIFO Sink",
		"Sink/PulseAudio",
		"Write stream to a virtual FIFO source on a PulseAudio server",
		"Avi Zimmerman <avi.zimmerman@gmail.com>",
	)
	cat.LogInfo("Adding sink pad template and properties to class")
	class.AddPadTemplate(staticSinkPadTemplate)
	class.InstallProperties(properties)
}

func (p *pulsefifosink) InstanceInit(self *glib.Object) {
	pad := gst.NewPadFromTemplate(staticSinkPadTemplate, "sink")
	gst.ToElement(self).AddPad(pad)
	p.state.SetSinkPad(pad)
}

func (p *pulsefifosink) GetProperty(self *glib.Object, id uint) *glib.Value {
	prop := properties[id]
	var localVal interface{}

	switch prop.Name() {
	case "server":
		localVal = p.settings.server
	case "device-name":
		localVal = p.settings.deviceName
	case "device-path":
		localVal = p.settings.devicePath
	}

	val, err := glib.GValue(localVal)
	if err != nil {
		gst.ToElement(self).Error(fmt.Sprintf("Could not convert %v to GValue", localVal), err)
		return nil
	}

	return val
}

func (p *pulsefifosink) SetProperty(self *glib.Object, id uint, value *glib.Value) {
	if p.state.running {
		gst.ToElement(self).Error("", errors.New("Cannot change element properties while running"))
		return
	}

	prop := properties[id]

	val, err := value.GoValue()
	if err != nil {
		gst.ToElement(self).Error(fmt.Sprintf("Could not coerce %v to go value", value), err)
	}

	switch prop.Name() {
	case "server":
		p.settings.server = val.(string)
	case "device-name":
		p.settings.deviceName = val.(string)
	case "device-path":
		p.settings.devicePath = val.(string)
	}
}

func (p *pulsefifosink) ChangeState(self *gst.Element, change gst.StateChange) gst.StateChangeReturn {
	switch change {

	case gst.StateChangeNullToReady:
		srvr := p.settings.server
		if srvr == "" {
			cat.LogInfo("Connecting to default PulseAudio server")
		} else {
			cat.LogInfo(fmt.Sprintf("Connecting to PulseAudio server on %s", srvr))
		}
		deviceManager, err := pa.NewDeviceManager(&pa.DeviceManagerOpts{PulseServer: srvr})
		if err != nil {
			self.Error("Could not connect to PulseAudio", err)
			return gst.StateChangeFailure
		}
		p.state.SetDeviceManager(deviceManager)

	case gst.StateChangeReadyToPaused:
		// Install a probe that will wait until we have fixed caps from the upstream peer
		p.state.GetSinkPad().AddProbe(gst.PadProbeTypePush, func(this *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
			return p.probeUntilFixated(self, this, info)
		})
		return gst.StateChangeAsync

	case gst.StateChangeReadyToNull:
		cat.LogInfo("Removing child elements")
		elems, err := gst.ToGstBin(self).GetElements()
		if err != nil {
			self.Error("Could not retrieve child elements", err)
			return gst.StateChangeFailure
		}
		for _, elem := range elems {
			if err := elem.SetState(gst.StateNull); err != nil {
				self.Error("Failed to change child element state", err)
				return gst.StateChangeFailure
			}
			if err := gst.ToGstBin(self).Remove(elem); err != nil {
				self.Error("Failed to remove child element from bin", err)
				return gst.StateChangeFailure
			}
		}
		// Tear down the device manager, this will also remove the created pipe source
		if devices := p.state.GetDeviceManager(); devices != nil {
			cat.LogInfo("Tearing down PA device manager")
			if err := devices.Destroy(); err != nil {
				self.Error("Could not tear down PA device manager", err)
			}
			p.state.SetDeviceManager(nil)
		}
		// If we created a temp directory remove it
		if p.state.GetIsTempDir() {
			cat.LogInfo("Removing temp directory")
			if err := os.RemoveAll(path.Dir(p.settings.devicePath)); err != nil {
				self.Error("Could not remove temp directory", err)
			}
		}
	}

	return gst.StateChangeSuccess
}

func (p *pulsefifosink) probeUntilFixated(self *gst.Element, this *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
	query := info.GetQuery()
	if query.Instance() == nil || query.Type() != gst.QueryCaps {
		return gst.PadProbeOK
	}
	// Get the caps from the query
	currentCaps := query.ParseCaps()

	// If the caps aren't fixed, continue probing
	if currentCaps == nil || currentCaps.IsAny() || !currentCaps.IsFixed() {
		return gst.PadProbeOK
	}

	// If the caps are fixed, we can go ahead and try to create a pulse pipe source
	// then link the elements to it.

	// Retrieve the values in the caps and convert to options for a new device
	values := currentCaps.GetStructureAt(0).Values()
	opts := p.capsValuesToPulseSourceOpts(self, values)
	if opts == nil {
		// we posted the error in the function below during conversion
		self.ContinueState(gst.StateChangeFailure)
		return gst.PadProbeRemove
	}
	cat.LogInfo(fmt.Sprintf("Creating new pipe-source with the following params: %+v", *opts))

	// Create the pipe source
	_, err := p.state.GetDeviceManager().AddSource(opts)
	if err != nil {
		// Post the error
		self.Error("Could not create PA fifo device", err)
		if p.state.GetIsTempDir() {
			if err := os.RemoveAll(path.Dir(p.settings.devicePath)); err != nil {
				self.Error("Failed to cleanup temp directory", err)
			}
		}
		self.ContinueState(gst.StateChangeFailure)
		return gst.PadProbeRemove
	}

	// Create our subelements
	cat.LogInfo("Setting up subelements")
	elements, err := gst.NewElementMany("queue", "filesink", "pulsesrc", "fakesink")
	if err != nil {
		self.Error("Could not create subelements", err)
		self.ContinueState(gst.StateChangeFailure)
		return gst.PadProbeRemove
	}
	queue, filesink, pulsesrc, fakesink := elements[0], elements[1], elements[2], elements[3]

	// Configure the filesink
	filesink.SetArg("location", p.settings.devicePath)
	filesink.SetArg("append", "true")
	filesink.SetArg("sync", "true")

	// Configure the pulsesrc and fakesink for pre-dumping the fifo
	if p.settings.server != "" {
		pulsesrc.SetArg("server", p.settings.server)
	}
	pulsesrc.SetArg("device", p.settings.deviceName)
	fakesink.SetArg("sync", "false")

	if err := gst.ToGstBin(self).AddMany(elements...); err != nil {
		self.Error("Could not add subelements to bin", err)
		self.ContinueState(gst.StateChangeFailure)
		return gst.PadProbeRemove
	}

	if err := queue.Link(filesink); err != nil {
		self.Error("Could not link queue to filesink", err)
		self.ContinueState(gst.StateChangeFailure)
		return gst.PadProbeRemove
	}
	if err := pulsesrc.Link(fakesink); err != nil {
		self.Error("Could not link queue to filesink", err)
		self.ContinueState(gst.StateChangeFailure)
		return gst.PadProbeRemove
	}

	ghostPad := gst.NewGhostPad("queue", queue.GetStaticPad("sink"))
	self.AddPad(ghostPad.Pad)

	peer := this.GetPeer()
	peer.Unlink(this)
	peer.Link(ghostPad.Pad)
	self.RemovePad(this)

	for _, elem := range elements {
		elem.SyncStateWithParent()
	}

	// Commit the state change and remove the probe
	self.ContinueState(gst.StateChangeSuccess)
	return gst.PadProbeRemove
}

func (p *pulsefifosink) capsValuesToPulseSourceOpts(self *gst.Element, values map[string]interface{}) *pa.SourceOpts {
	format, ok := values["format"]
	if !ok {
		format = "S16LE"
	}
	channels, ok := values["channels"]
	if !ok {
		channels = 1
	}
	rate, ok := values["rate"]
	if !ok {
		rate = 16000
	}
	if p.settings.devicePath == "" {
		dir, err := os.MkdirTemp("", "")
		if err != nil {
			self.Error("Could not get temp directory", err)
			return nil
		}
		p.state.SetIsTempDir(true)
		p.settings.devicePath = path.Join(dir, "src.fifo")
	}
	return &pa.SourceOpts{
		Name:         p.settings.deviceName,
		Description:  p.settings.deviceName,
		SampleFormat: strings.ToLower(format.(string)),
		SampleRate:   rate.(int),
		Channels:     channels.(int),
		FifoPath:     p.settings.devicePath,
	}
}
