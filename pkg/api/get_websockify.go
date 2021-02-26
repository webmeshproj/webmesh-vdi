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
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/proxyproto"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/lock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gorilla/websocket"
)

// swagger:operation GET /api/desktops/ws/{namespace}/{name}/display Desktops doWebsocket
// ---
// summary: Start an mTLS noVNC connection with the provided Desktop.
// description: Assumes the requesting client is a noVNC RFB object.
// parameters:
// - name: namespace
//   in: path
//   description: The namespace of the desktop session
//   type: string
//   required: true
// - name: name
//   in: path
//   description: The name of the desktop session
//   type: string
//   required: true
// - name: token
//   in: query
//   description: The X-Session-Token of the requesting client
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
func (d *desktopAPI) GetWebsockify(w http.ResponseWriter, r *http.Request) {
	lockName := fmt.Sprintf(
		"display-%s",
		strings.Replace(apiutil.GetNamespacedNameFromRequest(r).String(), "/", "-", -1),
	)
	labels := d.vdiCluster.GetComponentLabels("display-lock")
	labels[v1.ClientAddrLabel] = strings.Split(r.RemoteAddr, ":")[0] // Populated by ProxyHeaders handler wrapping the router
	sessionLock := lock.New(d.client, lockName, -1).WithLabels(labels)

	if err := sessionLock.Acquire(); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	defer func() {
		if err := sessionLock.Release(); err != nil {
			apiLogger.Error(err, "Failed to release lock on desktop display")
		}
	}()

	d.ServeWebsocketProxy(w, r, proxyproto.RequestTypeDisplay)
}

// swagger:operation GET /api/desktops/ws/{namespace}/{name}/audio Desktops doAudio
// ---
// summary: Retrieve the audio stream from the given desktop session.
// parameters:
// - name: namespace
//   in: path
//   description: The namespace of the desktop session
//   type: string
//   required: true
// - name: name
//   in: path
//   description: The name of the desktop session
//   type: string
//   required: true
// - name: token
//   in: query
//   description: The X-Session-Token of the requesting client. Can also be provided in the header.
//   type: string
//   required: false
// responses:
//   "UPGRADE": {}
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetWebsockifyAudio(w http.ResponseWriter, r *http.Request) {
	lockName := fmt.Sprintf(
		"audio-%s",
		strings.Replace(apiutil.GetNamespacedNameFromRequest(r).String(), "/", "-", -1),
	)
	labels := d.vdiCluster.GetComponentLabels("audio-lock")
	labels[v1.ClientAddrLabel] = strings.Split(r.RemoteAddr, ":")[0] // Populated by ProxyHeaders handler wrapping the router
	sessionLock := lock.New(d.client, lockName, -1).WithLabels(labels)

	if err := sessionLock.Acquire(); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	defer func() {
		if err := sessionLock.Release(); err != nil {
			apiLogger.Error(err, "Failed to release lock on desktop audio")
		}
	}()

	d.ServeWebsocketProxy(w, r, proxyproto.RequestTypeAudio)
}

var upgrader = &websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
	Subprotocols:      []string{"binary"},
	ReadBufferSize:    v1.WebsocketReadBufferSize,
	WriteBufferSize:   v1.WebsocketWriteBufferSize,
}

func (d *desktopAPI) ServeWebsocketProxy(w http.ResponseWriter, r *http.Request, rt proxyproto.RequestType) {
	proxy, err := d.getProxyClientForRequest(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiLogger.Info("Connecting to desktop proxy", "Path", r.URL.Path)

	var conn *proxyproto.Conn
	switch rt {
	case proxyproto.RequestTypeDisplay:
		conn, err = proxy.DisplayProxy()
	case proxyproto.RequestTypeAudio:
		conn, err = proxy.AudioProxy()
	}
	if err != nil {
		apiLogger.Error(err, "Error creating connection to proxy server")
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer conn.Close()

	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		apiLogger.Error(err, "Failed to upgrade the websocket connection")
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer wsconn.Close()

	client := apiutil.NewGorillaReadWriter(wsconn)
	ctx, cancel := context.WithCancel(context.Background())

	// Copy client connection to server
	go func() {
		defer cancel()
		if _, err := io.Copy(conn, client); err != nil {
			apiLogger.Error(err, "Error while copying stream from websocket connection to proxy")
		}
	}()

	// Copy server connection to the client
	go func() {
		defer cancel()
		if _, err := io.Copy(client, conn); err != nil {
			apiLogger.Error(err, "Error while copying stream from proxy to websocket connection")
		}
	}()

	// block until the context is finished
	for range ctx.Done() {
	}
}
