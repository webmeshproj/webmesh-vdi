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

// The main entrypoint to the kVDI App/API server.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/api"
	"github.com/kvdi/kvdi/pkg/util/common"
	"github.com/kvdi/kvdi/pkg/util/tlsutil"
)

var applogger = logf.Log.WithName("app")

func main() {
	var vdiCluster string
	var enableCORS bool
	flag.StringVar(&vdiCluster, "vdi-cluster", "", "The VDICluster this application is serving")
	flag.BoolVar(&enableCORS, "enable-cors", false, "Add CORS headers to requests")
	common.ParseFlagsAndSetupLogging()

	common.PrintVersion(applogger)

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		applogger.Error(err, "Failed to load kubernetes configuration")
		os.Exit(1)
	}

	// build the server
	srvr, err := newServer(cfg, vdiCluster, enableCORS)
	if err != nil {
		applogger.Error(err, "Failed to build the server router")
		os.Exit(1)
	}

	// serve
	applogger.Info(fmt.Sprintf("Starting VDI cluster frontend on :%d", v1.WebPort))
	if err := srvr.ListenAndServeTLS(tlsutil.ServerKeypair()); err != nil {
		applogger.Error(err, "Failed to start https server")
		os.Exit(1)
	}
}

// LogOutput is the object used to marshal log events to JSON.
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

func newServer(cfg *rest.Config, vdiCluster string, enableCORS bool) (*http.Server, error) {
	r := mux.NewRouter()
	// build the api router with our kubeconfig
	apiRouter, err := api.NewFromConfig(cfg, vdiCluster)
	if err != nil {
		return nil, err
	}
	// api routes
	r.PathPrefix("/api").Handler(apiRouter)
	// vue frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("/static/")))
	wrappedRouter := handlers.ProxyHeaders(
		handlers.CompressHandler(
			handlers.CustomLoggingHandler(os.Stdout, r, formatLog),
		),
	)
	if enableCORS {
		wrappedRouter = handlers.CORS()(wrappedRouter)
	}
	return &http.Server{
		Handler: wrappedRouter,
		Addr:    fmt.Sprintf(":%d", v1.WebPort),
		// TODO: make these configurable (currently high for large dir transfers)
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}, nil
}
