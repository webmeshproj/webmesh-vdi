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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/api"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
)

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
	// build the api router with our kubeconfig
	apiRouter, err := api.NewFromConfig(cfg, vdiCluster)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	// api routes
	r.PathPrefix("/api").Handler(apiRouter)
	// vue frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("/static")))

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
