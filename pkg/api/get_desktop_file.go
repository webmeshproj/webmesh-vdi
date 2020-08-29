package api

import (
	"net/http"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
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
