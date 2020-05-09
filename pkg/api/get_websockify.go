package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
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
	d.ServeWebsocketProxy(w, r)
}

func (d *desktopAPI) ServeWebsocketProxy(w http.ResponseWriter, r *http.Request) {
	endpointURL, err := d.getEndpointURL(r)
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
