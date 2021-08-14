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

	"github.com/kennygrant/sanitize"
	"github.com/kvdi/kvdi/pkg/proxyproto"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation PUT /api/desktops/fs/{namespace}/{name}/put Desktops putDesktopFile
// ---
// summary: Uploads a file to a desktop session.
// consumes:
// - multipart/form-data
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
// - in: formData
//   name: file
//   type: file
//   description: The file to upload.
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) PutDesktopFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	proxy, err := d.getProxyClientForRequest(r)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	if err := proxy.PutFile(&proxyproto.FPutRequest{
		Name: sanitize.BaseName(handler.Filename),
		Size: handler.Size,
		Body: file,
	}); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteOK(w)
}
