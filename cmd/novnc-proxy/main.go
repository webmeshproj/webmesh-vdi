// The main entrypoint for the novnc-proxy which provides an mTLS websocket server in front of display and audio streams.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/tinyzimmer/kvdi/pkg/util/audio"
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
var log = logf.Log.WithName("novnc_proxy")

// vnc configurations
var vncAddr string
var userID string
var vncConnectProto, vncConnectAddr string

// main application entry point
func main() {

	// parse flags and setup logging
	pflag.CommandLine.StringVar(&vncAddr, "vnc-addr", "unix:///var/run/kvdi/display.sock", "The tcp or unix-socket address of the vnc server")
	pflag.CommandLine.StringVar(&userID, "user-id", "9000", "The user ID directory in /run/usr where sockets are located")
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

	log.Info(fmt.Sprintf("Starting mTLS enabled novnc proxy on :%d", v1.WebPort))
	if err := server.ListenAndServeTLS(tlsutil.ServerKeypair()); err != nil {
		log.Error(err, "Failed to start https server")
		os.Exit(1)
	}
}

// newServer builds the novnc proxy server
func newServer() (*http.Server, error) {
	tlsConfig, err := tlsutil.NewServerTLSConfig()
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	// The websockify route is in charge of proxying noVNC conncetions to the local
	// VNC socket. This route is pretty bulletproof.
	r.Path("/api/desktops/{namespace}/{name}/websockify").Handler(&websocket.Server{
		Handshake: func(c *websocket.Config, r *http.Request) error {
			return nil
		},
		Handler: func(wsconn *websocket.Conn) {
			log.Info(fmt.Sprintf("Received display proxy request, connecting to %s", vncAddr))
			conn, err := net.Dial(vncConnectProto, vncConnectAddr)

			if err != nil {
				log.Error(err, "Failed to connect to display server")
				wsconn.Close()
				return
			}

			log.Info("Connection established, proxying display")

			wsconn.PayloadType = websocket.BinaryFrame

			go func() {
				if _, err := io.Copy(conn, wsconn); err != nil {
					log.Error(err, "Error while copying stream from websocket connection to display socket")
				}
			}()
			go func() {
				if _, err := io.Copy(wsconn, conn); err != nil {
					log.Error(err, "Error while copying stream from display socket to websocket connection")
				}
			}()

			select {}
		},
	})

	// This route creates a recorder on the local pulseaudio sink and ships
	// the data back to the client over a websocket.
	r.Path("/api/desktops/{namespace}/{name}/wsaudio").Handler(&websocket.Server{
		Handshake: func(c *websocket.Config, r *http.Request) error {
			return nil
		},
		Handler: func(wsconn *websocket.Conn) {
			log.Info(fmt.Sprintf("Received validated proxy request, connecting to audio stream"))

			wsconn.PayloadType = websocket.BinaryFrame

			audioBuffer := audio.NewBuffer(log, userID, audio.BufferTypeGST)

			if err := audioBuffer.Start(audio.CodecOpus); err != nil {
				log.Error(err, "Error setting up audio buffer")
				return
			}

			defer func() {
				if err := audioBuffer.Close(); err != nil {
					log.Error(err, "Error closing audio buffer")
				}
			}()

			go func() {

				if _, err := io.Copy(wsconn, audioBuffer); err != nil {
					log.Error(err, "Error while copying from audio stream to websocket connection")
				}
			}()

			if err := audioBuffer.Wait(); err != nil {
				if err := audioBuffer.Error(); err != nil {
					log.Error(err, "Error while streaming audio")
					if _, err := wsconn.Write(append([]byte(err.Error()), []byte("\n")...)); err != nil {
						log.Error(err, "Failed to write error to websocket client")
					}
				}
			}

			if err := wsconn.Close(); err != nil {
				log.Error(err, "Error closing websocket connection")
			}

			log.Info("Finishing proxying audio stream")
		},
	})

	wrapped := handlers.CustomLoggingHandler(os.Stdout, r, formatLog)

	return &http.Server{
		Handler:      wrapped,
		Addr:         fmt.Sprintf(":%d", v1.WebPort),
		TLSConfig:    tlsConfig,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, nil
}

// LogOutput represents a log message
type LogOutput struct {
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"statusCode"`
	Size       int       `json:"size"`
	RemoteHost string    `json:"remoteHost"`
}

func formatLog(writer io.Writer, params handlers.LogFormatterParams) {
	host, _, err := net.SplitHostPort(params.Request.RemoteAddr)
	if err != nil {
		host = params.Request.RemoteAddr
	}
	if out, err := json.Marshal(&LogOutput{
		Time:       params.TimeStamp,
		Method:     params.Request.Method,
		Path:       params.URL.Path,
		StatusCode: params.StatusCode,
		RemoteHost: host,
		Size:       params.Size,
	}); err == nil {
		if _, err := writer.Write(append(out, []byte("\n")...)); err != nil {
			fmt.Println(string(out))
		}
	}
}
