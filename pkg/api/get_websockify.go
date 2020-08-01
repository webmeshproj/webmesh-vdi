package api

import (
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/lock"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// swagger:operation GET /api/desktops/{namespace}/{name}/websockify Desktops doWebsocket
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
	labels[v1.ClientAddrLabel] = getClientIP(r)
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

	d.ServeWebsocketProxy(w, r)
}

// swagger:operation GET /api/desktops/{namespace}/{name}/wsaudio Desktops doAudio
// ---
// summary: Retrive the audio stream from the given desktop session.
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
	labels[v1.ClientAddrLabel] = getClientIP(r)
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

	d.ServeWebsocketProxy(w, r)
}

func (d *desktopAPI) ServeWebsocketProxy(w http.ResponseWriter, r *http.Request) {
	endpointURL, err := d.getDesktopWebsocketURL(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	clientTLSConfig, err := tlsutil.NewClientTLSConfig()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiLogger.Info("Starting new websocket proxy", "Host", endpointURL, "Path", r.URL.Path)
	proxy := websocketproxy.NewProxy(endpointURL)
	proxy.Dialer = &websocket.Dialer{
		TLSClientConfig: clientTLSConfig,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	proxy.Upgrader = &upgrader
	proxy.ServeHTTP(w, r)
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	// strip port if present
	return strings.Split(ip, ":")[0]
}
