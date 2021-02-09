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
