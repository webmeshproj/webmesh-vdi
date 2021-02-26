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

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/proxyproto"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

func getLocalPathFromRequest(path string) (string, error) {
	if _, err := os.Stat(v1.DesktopHomeMntPath); err != nil {
		return "", errors.New("File transfer is disabled for this desktop session")
	}

	fPath := filepath.Join(v1.DesktopHomeMntPath, path)
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

func (p *Server) logConnectionMetrics(proxyType string, conn *proxyproto.Conn) chan struct{} {
	st := make(chan struct{})
	logger := p.log.WithValues("Connection", proxyType)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-st:
				logger.Info("Connection is closing")
				return
			case <-ticker.C:
				logger.Info("Connection is alive", "BytesSent", conn.BytesSentCount(), "BytesReceived", conn.BytesRecvdCount())
			}
		}
	}()
	return st
}
