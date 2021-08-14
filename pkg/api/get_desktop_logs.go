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

package api

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/kvdi/kvdi/pkg/util/errors"
	"github.com/kvdi/kvdi/pkg/util/k8sutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"golang.org/x/net/websocket"
	corev1 "k8s.io/api/core/v1"
)

// swagger:operation GET /api/desktops/{namespace}/{name}/logs/{container} Desktops getLogs
// ---
// summary: Retrieve the logs for a container in a desktop session.
// parameters:
// - name: namespace
//   in: path
//   description: The namespace of the desktop session.
//   type: string
//   required: true
// - name: name
//   in: path
//   description: The name of the desktop session.
//   type: string
//   required: true
// - name: container
//   in: path
//   description: The container to retrieve logs for. Can be 'kvdi-proxy' or 'desktop'.
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/getLogsResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopLogs(w http.ResponseWriter, r *http.Request) {
	pod, err := d.getDesktopPodForRequest(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	container := apiutil.GetContainerFromRequest(r)
	logRdr := k8sutil.NewLogFollower(pod, container)
	if err := logRdr.Stream(false); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer logRdr.Close()
	if _, err := io.Copy(w, logRdr); err != nil {
		apiLogger.Error(err, "Error writing log stream to the HTTP response")
	}
}

// Desktop logs response
// swagger:response getLogsResponse
type swaggerGetLogsResponse struct {
	// in:body
	Body string
}

// swagger:operation GET /api/desktops/ws/{namespace}/{name}/logs/{container} Desktops getLogsWebsocket
// ---
// summary: Follow the logs for a desktop over a websocket.
// parameters:
// - name: namespace
//   in: path
//   description: The namespace of the desktop session.
//   type: string
//   required: true
// - name: name
//   in: path
//   description: The name of the desktop session.
//   type: string
//   required: true
// - name: container
//   in: path
//   description: The container to retrieve logs for. Can be 'kvdi-proxy' or 'desktop'.
//   type: string
//   required: true
// - name: token
//   in: query
//   description: The X-Session-Token of the requesting client.
//   type: string
//   required: true
// responses:
//   "UPGRADE": {}
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopLogsWebsocket(wsconn *websocket.Conn) {
	defer wsconn.Close()

	pod, err := d.getDesktopPodForRequest(wsconn.Request())
	if err != nil {
		var apiError *errors.APIError
		if client.IgnoreNotFound(err) == nil {
			apiError = errors.ToAPIError(err, errors.NotFound)
		} else {
			apiError = errors.ToAPIError(err, errors.ServerError)
		}
		if _, werr := wsconn.Write(apiError.JSON()); werr != nil {
			apiLogger.Error(err, "Error retrieving pod for request")
			apiLogger.Error(werr, "Failed to write error to websocket connection")
		}
		return
	}
	container := apiutil.GetContainerFromRequest(wsconn.Request())
	logRdr := k8sutil.NewLogFollower(pod, container)
	if err := logRdr.Stream(true); err != nil {
		if _, werr := wsconn.Write(errors.ToAPIError(err, errors.ServerError).JSON()); werr != nil {
			apiLogger.Error(err, "Error retrieving logs from pod")
			apiLogger.Error(werr, "Failed to write error to websocket connection")
		}
		return
	}

	defer logRdr.Close()

	buf := bufio.NewReader(logRdr)

	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// EOF in this context means we are still waiting for data on the stream.
				// Sleep and continue.
				time.Sleep(time.Second)
				continue
			}
			if _, werr := wsconn.Write(errors.ToAPIError(err, errors.ServerError).JSON()); werr != nil {
				apiLogger.Error(err, "Error occured while reading from log reader")
				apiLogger.Error(werr, "Failed to write error to websocket connection")
			}
			return
		}
		if _, err := wsconn.Write(line); err != nil {
			if errors.IsBrokenPipeError(err) {
				apiLogger.Info("Client has disconnected, finishing log stream")
				return
			}
			apiLogger.Error(err, "Error while writing log event to websocket")
			return
		}
	}

}

func (d *desktopAPI) getDesktopPodForRequest(r *http.Request) (*corev1.Pod, error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &corev1.Pod{}
	return found, d.client.Get(context.TODO(), nn, found)
}
