package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:operation DELETE /api/sessions/{namespace}/{name} Desktops deleteSession
// ---
// summary: Destroys the provided desktop session.
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
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "500":
//     "$ref": "#/responses/error"
func (d *desktopAPI) DeleteDesktopSession(w http.ResponseWriter, r *http.Request) {
	nn := getNamespacedNameFromRequest(r)
	found := &v1alpha1.Desktop{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := d.client.Delete(context.TODO(), found); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
