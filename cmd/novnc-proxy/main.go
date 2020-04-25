package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"golang.org/x/net/websocket"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("novnc_proxy")
var vncAddr string

var vncConnectProto, vncConnectAddr string

func main() {
	pflag.CommandLine.StringVar(&vncAddr, "vnc-addr", "tcp://127.0.0.1:5900", "The tcp or unix-socket address of the vnc server")
	common.ParseFlagsAndSetupLogging()

	if strings.HasPrefix(vncAddr, "tcp://") {
		vncConnectProto = "tcp"
		vncConnectAddr = strings.TrimPrefix(vncAddr, "tcp://")
	} else if strings.HasPrefix(vncAddr, "unix://") {
		vncConnectProto = "unix"
		vncConnectAddr = strings.TrimPrefix(vncAddr, "unix://")
	} else {
		log.Info(fmt.Sprintf("%s is an invalid vnc address", vncAddr))
		os.Exit(1)
	}

	server, err := newServer()
	if err != nil {
		log.Error(err, "Failed to create https server")
		os.Exit(1)
	}

	log.Info(fmt.Sprintf("Starting mTLS enabled novnc proxy on :%d", v1alpha1.WebPort))
	if err := server.ListenAndServeTLS(tlsutil.ServerKeypair()); err != nil {
		log.Error(err, "Failed to start https server")
		os.Exit(1)
	}
}

func newServer() (*http.Server, error) {
	tlsConfig, err := tlsutil.NewServerTLSConfig()
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	r.PathPrefix("/").Handler(&websocket.Server{
		Handshake: func(c *websocket.Config, r *http.Request) error {
			return nil
		},
		Handler: func(wsconn *websocket.Conn) {
			log.Info(fmt.Sprintf("Received validated proxy request, connecting to %s", vncAddr))
			conn, err := net.Dial(vncConnectProto, vncConnectAddr)

			if err != nil {
				log.Error(err, "Failed to connect to VNC server")
				wsconn.Close()
				return

			}

			log.Info("Connection established, proxying vnc session")

			wsconn.PayloadType = websocket.BinaryFrame

			go func() {
				if _, err := io.Copy(conn, wsconn); err != nil {
					log.Error(err, "Error while copying stream from websocket connection to VNC socket")
				}
			}()
			go func() {
				if _, err := io.Copy(wsconn, conn); err != nil {
					log.Error(err, "Error while copying stream from VNC socket to websocket connection")
				}
			}()

			select {}
		},
	})

	return &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%d", v1alpha1.WebPort),
		TLSConfig:    tlsConfig,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, nil
}
