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
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tinyzimmer/kvdi/pkg/proxyproto"
	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
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
	proxy, err := d.getProxyClientForRequest(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	path := getPathFromRequest(r)
	res, err := proxy.StatFile(&proxyproto.FStatRequest{
		Path: path,
	})
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer res.Close()
	if _, err := io.Copy(w, res); err != nil {
		apiLogger.Error(err, "Error copying proxy response to client")
	}
}

// File stat response
// swagger:response statDesktopFileResponse
type swaggerStatDesktopFileResponse struct {
	// in:body
	Body types.StatDesktopFileResponse
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
	proxy, err := d.getProxyClientForRequest(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	path := getPathFromRequest(r)
	res, err := proxy.GetFile(&proxyproto.FGetRequest{
		Path: path,
	})
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer res.Body.Close()

	fileSizeStr := strconv.FormatInt(res.Size, 10)

	w.Header().Set("Content-Length", fileSizeStr)
	w.Header().Set("Content-Type", res.Type)
	w.Header().Set("Content-Disposition", "attachment; filename="+res.Name)
	w.Header().Set("X-Suggested-Filename", res.Name)
	w.Header().Set("X-Decompressed-Content-Length", fileSizeStr)
	w.WriteHeader(http.StatusOK)

	// Copy the file contents to the response
	if _, err := io.Copy(w, res.Body); err != nil {
		apiLogger.Error(err, "Failed to copy file contents to response buffer")
	}
}

func getPathFromRequest(r *http.Request) string {
	pathPrefix := apiutil.GetGorillaPath(r)
	pathPrefix = strings.Replace(pathPrefix, "{name}", mux.Vars(r)["name"], 1)
	pathPrefix = strings.Replace(pathPrefix, "{namespace}", mux.Vars(r)["namespace"], 1)
	return strings.TrimPrefix(r.URL.Path, pathPrefix)
}
