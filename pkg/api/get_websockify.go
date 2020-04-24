package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	corev1 "k8s.io/api/core/v1"
)

var clientTLSConfig *tls.Config

func init() {
	var err error
	clientTLSConfig, err = tlsutil.NewClientTLSConfig()
	if err != nil {
		panic(err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// swagger:operation GET /api/websockify/{namespace}/{name} Desktops doWebsocket
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
//   "500":
//     "$ref": "#/responses/error"
func (d *desktopAPI) mtlsWebsockify(w http.ResponseWriter, r *http.Request) {
	endpointURL, err := d.getEndpointURL(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiLogger.Info(fmt.Sprintf("Starting new mTLS websocket proxy with %s", endpointURL))
	proxy := websocketproxy.NewProxy(endpointURL)
	proxy.Dialer = &websocket.Dialer{
		TLSClientConfig: clientTLSConfig,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	proxy.Upgrader = &upgrader
	proxy.ServeHTTP(w, r)
}

func (d *desktopAPI) getEndpointURL(r *http.Request) (*url.URL, error) {
	nn := getNamespacedNameFromRequest(r)
	// url, _ := url.Parse(fmt.Sprintf("wss://%s.%s.%s:%d", nn.Name, nn.Name, nn.Namespace, v1alpha1.WebPort))
	// return url
	found := &corev1.Service{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return nil, err
	}
	return url.Parse(fmt.Sprintf("wss://%s:%d", found.Spec.ClusterIP, v1alpha1.WebPort))
}
