package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
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
}

func mtlsWebsockify(w http.ResponseWriter, r *http.Request) {

	endpointURL := getEndpointURL(r)
	applogger.Info(fmt.Sprintf("Starting new mTLS websocket proxy with %s", endpointURL))
	proxy := websocketproxy.NewProxy(endpointURL)
	proxy.Dialer = &websocket.Dialer{
		TLSClientConfig: clientTLSConfig,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	proxy.Upgrader = &upgrader
	proxy.ServeHTTP(w, r)
}

func getEndpointURL(r *http.Request) *url.URL {
	vars := mux.Vars(r)
	endpoint := vars["endpoint"]
	url, _ := url.Parse(fmt.Sprintf("wss://%s:%d", endpoint, v1alpha1.WebPort))
	return url
}
