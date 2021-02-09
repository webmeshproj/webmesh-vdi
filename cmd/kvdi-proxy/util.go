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
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func getLocalPathFromRequest(r *http.Request) (path string, err error) {
	if _, err := os.Stat(v1.DesktopHomeMntPath); err != nil {
		return "", errors.New("File transfer is disabled for this desktop session")
	}

	// build out the path prefix to strip from the request URL
	pathPrefix := apiutil.GetGorillaPath(r)
	pathPrefix = strings.Replace(pathPrefix, "{name}", mux.Vars(r)["name"], 1)
	pathPrefix = strings.Replace(pathPrefix, "{namespace}", mux.Vars(r)["namespace"], 1)

	fPath := filepath.Join(v1.DesktopHomeMntPath, strings.TrimPrefix(r.URL.Path, pathPrefix))
	absPath, err := filepath.Abs(fPath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, v1.DesktopHomeMntPath) {
		// requestor tried to traverse outside the user's home directory (into proxy root fs)
		return "", fmt.Errorf("%s is outside the user's home directory", fPath)
	}
	return absPath, nil
}

func logWatcherMetrics(proxyType string, watcher *apiutil.WebsocketWatcher) chan struct{} {
	st := make(chan struct{})
	logger := log.WithValues("Connection", proxyType)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-st:
				logger.Info("Connection is closing")
				return
			case <-ticker.C:
				logger.Info("Connection is alive", "BytesSent", watcher.BytesSentCount(), "BytesReceived", watcher.BytesRecvdCount())
			}
		}
	}()
	return st
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
