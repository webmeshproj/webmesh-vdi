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

// The main entrypoint for the novnc-proxy which provides an mTLS websocket server in front of display and audio streams.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Uncomment this and see its usage below to enable server side audio buffering
// const audioBufferSize = 8 * 1024

// our local logger instance
var log = logf.Log.WithName("kvdi_proxy")

// vnc/audio configurations
var vncAddr string
var userID int
var pulseServer string
var vncConnectProto, vncConnectAddr string

// main application entry point
func main() {

	// parse flags and setup logging
	flag.StringVar(&vncAddr, "vnc-addr", "unix:///var/run/kvdi/display.sock", "The tcp or unix-socket address of the vnc server")
	flag.IntVar(&userID, "user-id", 9000, "The ID of the main user in the desktop container, used for chown operations")
	flag.StringVar(&pulseServer, "pulse-server", "", "The socket where pulseaudio is accepting connections. Defaults to /run/user/<userID>/pulse/native")

	common.ParseFlagsAndSetupLogging()
	common.PrintVersion(log)

	// Set the location of our vnc socket appropriatly
	if strings.HasPrefix(vncAddr, "tcp://") {
		vncConnectProto = "tcp"
		vncConnectAddr = strings.TrimPrefix(vncAddr, "tcp://")
	} else if strings.HasPrefix(vncAddr, "unix://") {
		vncConnectProto = "unix"
		vncConnectAddr = strings.TrimPrefix(vncAddr, "unix://")
	} else {
		// Should never happen as the manager is usually in charge of us
		log.Info(fmt.Sprintf("%s is an invalid vnc address", vncAddr))
		os.Exit(1)
	}

	// Populate the default pulseserver path if not set on the command line
	if pulseServer == "" {
		pulseServer = fmt.Sprintf("/run/user/%d/pulse/native", userID)
	}

	// build and run the server
	server, err := newServer()
	if err != nil {
		log.Error(err, "Failed to create https server")
		os.Exit(1)
	}

	log.Info(fmt.Sprintf("Starting kvdi proxy on :%d", v1.WebPort))
	if err := server.ListenAndServeTLS(tlsutil.ServerKeypair()); err != nil {
		log.Error(err, "Failed to start https server")
		os.Exit(1)
	}
}

// newServer builds the novnc proxy server
func newServer() (*http.Server, error) {
	r := mux.NewRouter()

	// The websockify route is in charge of proxying noVNC conncetions to the local
	// VNC socket. This route is pretty bulletproof.
	r.Path("/api/desktops/ws/{namespace}/{name}/display").Handler(&websocket.Server{
		Handshake: wsHandshake,
		Handler:   websockifyHandler,
	})

	// This route creates a recorder on the local pulseaudio sink and ships
	// the data back to the client over a websocket.
	r.Path("/api/desktops/ws/{namespace}/{name}/audio").Handler(&websocket.Server{
		Handshake: wsHandshake,
		Handler:   wsAudioHandler,
	})

	// This route is for doing a stat of files in the user's home directory when
	// enabled in the DesktopTemplate.
	r.PathPrefix("/api/desktops/fs/{namespace}/{name}/stat/").HandlerFunc(statFileHandler)

	// This route is for downloading a file from the user's home directory when
	// enabled in the DesktopTemplate.
	r.PathPrefix("/api/desktops/fs/{namespace}/{name}/get/").HandlerFunc(downloadFileHandler)

	// This route is for uploading a file to the user's home directory when enabled in the
	// DesktopTemplate.
	r.PathPrefix("/api/desktops/fs/{namespace}/{name}/put").HandlerFunc(uploadFileHandler)

	wrapped := handlers.CustomLoggingHandler(os.Stdout, r, formatLog)

	tlsConfig, err := tlsutil.NewServerTLSConfig()
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Handler:   wrapped,
		Addr:      fmt.Sprintf(":%d", v1.WebPort),
		TLSConfig: tlsConfig,
		// TODO: make these configurable (currently high for large dir transfers)
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}, nil
}
