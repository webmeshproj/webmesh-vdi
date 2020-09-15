package api

import (
	"bufio"
	"context"
	"io"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"
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
	logRdr, err := k8sutil.GetPodLogs(pod, container, false)
	if err != nil {
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
		if _, werr := wsconn.Write(errors.ToAPIError(err).JSON()); werr != nil {
			apiLogger.Error(err, "Error retrieving pod for request")
			apiLogger.Error(werr, "Failed to write error to websocket connection")
		}
		return
	}
	container := apiutil.GetContainerFromRequest(wsconn.Request())
	logRdr, err := k8sutil.GetPodLogs(pod, container, true)
	if err != nil {
		if _, werr := wsconn.Write(errors.ToAPIError(err).JSON()); werr != nil {
			apiLogger.Error(err, "Error retrieving logs from pod")
			apiLogger.Error(werr, "Failed to write error to websocket connection")
		}
		return
	}

	defer logRdr.Close()

	scanner := bufio.NewScanner(logRdr)

	for scanner.Scan() {
		line := scanner.Bytes()
		if _, err := wsconn.Write(append(line, []byte("\n")...)); err != nil {
			if errors.IsBrokenPipeError(err) {
				return
			}
			apiLogger.Error(err, "Error while writing log event to websocket")
		}
	}

	if err := scanner.Err(); err != io.EOF && err != nil {
		if _, werr := wsconn.Write(errors.ToAPIError(err).JSON()); werr != nil {
			apiLogger.Error(err, "Error occured while scanning log reader")
			apiLogger.Error(werr, "Failed to write error to websocket connection")
		}
		return
	}
}

func (d *desktopAPI) getDesktopPodForRequest(r *http.Request) (*corev1.Pod, error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &corev1.Pod{}
	return found, d.client.Get(context.TODO(), nn, found)
}
