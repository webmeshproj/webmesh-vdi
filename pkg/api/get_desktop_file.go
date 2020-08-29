package api

import (
	"bufio"
	"io"
	"net/http"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation GET /api/desktops/fs/{namespace}/{name}/stat/{fpath} Desktops statDesktopFile
// ---
// summary: Retrieve filesystem info for the given path inside a desktop session.
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
// - name: fpath
//   in: path
//   description: The path to retrieve information about
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/statDesktopFileResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetStatDesktopFile(w http.ResponseWriter, r *http.Request) {
	d.serveHTTPProxy(w, r)
}

// File stat response
// swagger:response statDesktopFileResponse
type swaggerStatDesktopFileResponse struct {
	// in:body
	Body v1.StatDesktopFileResponse
}

// swagger:operation GET /api/desktops/fs/{namespace}/{name}/get/{fpath} Desktops downloadDesktopFile
// ---
// summary: Download the given file from a desktop session.
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
// - name: fpath
//   in: path
//   description: The file path to download
//   type: string
//   required: true
// responses:
//   "200":
//     content:
//       "application/octet-stream":
//         type: string
//         format: binary
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDownloadDesktopFile(w http.ResponseWriter, r *http.Request) {
	d.serveHTTPProxy(w, r)
}

func (d *desktopAPI) serveHTTPProxy(w http.ResponseWriter, r *http.Request) {
	desktopHost, err := d.getDesktopWebHost(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	// Overwrite the request object host to point to the desktop container
	u := r.URL
	u.Scheme = "https"
	u.Host = desktopHost

	// Buld a request from the source
	req, err := http.NewRequest(r.Method, u.String(), bufio.NewReader(r.Body))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// Build an HTTP client
	clientTLSConfig, err := tlsutil.NewClientTLSConfig()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientTLSConfig,
		},
	}

	// Do the request
	resp, err := httpClient.Do(req)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	defer resp.Body.Close()

	// copy the response from the proxy to the requestor
	w.WriteHeader(resp.StatusCode)
	for hdr, val := range resp.Header {
		w.Header().Add(hdr, strings.Join(val, ";"))
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		apiLogger.Error(err, "Error copying response body from desktop proxy")
	}

}
