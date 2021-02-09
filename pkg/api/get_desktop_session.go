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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"golang.org/x/net/websocket"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation GET /api/sessions/{namespace}/{name} Sessions getSession
// ---
// summary: Retrieve the status of the requested desktop session.
// description: Details include the PodPhase and CRD status.
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
// responses:
//   "200":
//     "$ref": "#/responses/getSessionResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopSessionStatus(w http.ResponseWriter, r *http.Request) {
	desktop, err := d.getDesktopForRequest(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(fmt.Errorf("No desktop session %s found", apiutil.GetNamespacedNameFromRequest(r).String()), w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(toReturnStatus(desktop), w)
}

// Session status response
// swagger:response getSessionResponse
type swaggerGetSessionResponse struct {
	// in:body
	Body map[string]interface{}
}

// swagger:operation GET /api/desktops/ws/{namespace}/{name}/status Desktops getSessionStatusWs
// ---
// summary: Retrieve status updates of the requested desktop session over a websocket.
// description: Details include the PodPhase and CRD status.
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
// responses:
//   "UPGRADE": {}
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopSessionStatusWebsocket(conn *websocket.Conn) {
	defer conn.Close()

	ticker := time.NewTicker(time.Duration(2) * time.Second)
	for range ticker.C {

		desktop, err := d.getDesktopForRequest(conn.Request())
		if err != nil {
			if _, err := conn.Write(errors.ToAPIError(err).JSON()); err != nil {
				apiLogger.Error(err, "Failed to write error to websocket connection")
				return
			}
			if client.IgnoreNotFound(err) == nil {
				// If the desktop doesn't exist, we should give up entirely.
				// Other api errors are worth letting the client retry.
				return
			}
		}
		st := toReturnStatus(desktop)
		if _, err := conn.Write(st.JSON()); err != nil {
			apiLogger.Error(err, "Failed to write status to websocket connection")
			return
		}

		if st.Running && st.PodPhase == corev1.PodRunning {
			// we are done here, the client shouldn't need anything else
			return
		}

	}
}

func (d *desktopAPI) getDesktopForRequest(r *http.Request) (*desktopsv1.Session, error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &desktopsv1.Session{}
	return found, d.client.Get(context.TODO(), nn, found)
}

type desktopStatus struct {
	Running  bool            `json:"running"`
	PodPhase corev1.PodPhase `json:"podPhase"`
}

func toReturnStatus(desktop *desktopsv1.Session) *desktopStatus {
	return &desktopStatus{
		Running:  desktop.Status.Running,
		PodPhase: desktop.Status.PodPhase,
	}
}

func (d *desktopStatus) JSON() []byte {
	out, _ := json.Marshal(d)
	return out
}
