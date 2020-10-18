// The main entrypoint for the novnc-proxy which provides an mTLS websocket server in front of display and audio streams.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"golang.org/x/net/websocket"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Uncomment this and see its usage below to enable server side audio buffering
// const audioBufferSize = 8 * 1024

// our local logger instance
var log = logf.Log.WithName("kvdi_proxy")

// vnc configurations
var vncAddr string
var userID int
var vncConnectProto, vncConnectAddr string

// main application entry point
func main() {

	// parse flags and setup logging
	pflag.CommandLine.StringVar(&vncAddr, "vnc-addr", "unix:///var/run/kvdi/display.sock", "The tcp or unix-socket address of the vnc server")
	pflag.CommandLine.IntVar(&userID, "user-id", 9000, "The ID of the main user in the desktop container")
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
