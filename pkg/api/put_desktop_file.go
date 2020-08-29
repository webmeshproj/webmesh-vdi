package api

import "net/http"

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
	d.serveHTTPProxy(w, r)
}
