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

// The main entrypoint for the kvdi-proxy which provides an mTLS TCP server in front of desktop instances.
// The server provides access to display/audio streams and filesystem operations.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	proxyserver "github.com/kvdi/kvdi/pkg/proxyproto/server"
	"github.com/kvdi/kvdi/pkg/util/common"
)

// TODO: clean this all up

var (
	log = logf.Log.WithName("kvdi_proxy")

	listenHost string

	userID                                  int
	pulseServer                             string
	displayAddr                             string
	displayConnectProto, displayConnectAddr string

	monitorDeviceName    = "kvdi"
	monitorDescription   = "kvdi-playback"
	micDeviceName        = "virtmic"
	micDeviceDescription = "kvdi-microphone"
	micDevicePath        = filepath.Join(v1.DesktopRunDir, "mic.fifo")
	micDeviceFormat      = "s16le"
	micDeviceChannels    = 1
	micDeviceSampleRate  = 16000
)

// main application entry point
func main() {

	// parse flags and setup logging
	flag.StringVar(&listenHost, "listen", "0.0.0.0", "The address to listen for connections on")
	flag.StringVar(&displayAddr, "display-addr", "unix:///var/run/kvdi/display.sock", "The tcp or unix-socket address of the display server")
	flag.IntVar(&userID, "user-id", 9000, "The ID of the main user in the desktop container, used for chown operations")
	flag.StringVar(&pulseServer, "pulse-server", "", "The socket where pulseaudio is accepting connections. Defaults to /run/user/<userID>/pulse/native")
	common.ParseFlagsAndSetupLogging()
	common.PrintVersion(log)

	// Set the location of our vnc socket appropriatly
	if strings.HasPrefix(displayAddr, "tcp://") {
		displayConnectProto = "tcp"
		displayConnectAddr = strings.TrimPrefix(displayAddr, "tcp://")
	} else if strings.HasPrefix(displayAddr, "unix://") {
		displayConnectProto = "unix"
		displayConnectAddr = strings.TrimPrefix(displayAddr, "unix://")
	} else {
		// Should never happen as the manager is usually in charge of us
		log.Info(fmt.Sprintf("%s is an invalid display address", displayAddr))
		os.Exit(1)
	}

	// Populate the default pulseserver path if not set on the command line
	if pulseServer == "" {
		pulseServer = fmt.Sprintf("/run/user/%d/pulse/native", userID)
	}

	// build and run the server

	server := proxyserver.New(log, listenHost, v1.WebPort, &proxyserver.ProxyOpts{
		FSUserID:                   userID,
		DisplayAddress:             displayConnectAddr,
		DisplayProto:               displayConnectProto,
		PulseServer:                pulseServer,
		PlaybackDeviceName:         monitorDeviceName,
		PlaybackSampleRate:         24000, // TODO
		PlaybackDeviceDescription:  monitorDescription,
		RecordingDeviceName:        micDeviceName,
		RecordingDeviceDescription: micDeviceDescription,
		RecordingDevicePath:        micDevicePath,
		RecordingDeviceFormat:      micDeviceFormat,
		RecordingDeviceSampleRate:  micDeviceSampleRate,
		RecordingDeviceChannels:    micDeviceChannels,
	})

	if err := server.ListenAndServe(); err != nil {
		log.Error(err, "Error running proxy server")
		os.Exit(1)
	}
}
