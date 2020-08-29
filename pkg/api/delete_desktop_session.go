package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation DELETE /api/sessions/{namespace}/{name} Sessions deleteSession
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
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) DeleteDesktopSession(w http.ResponseWriter, r *http.Request) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &v1alpha1.Desktop{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(fmt.Errorf("No desktop session %s found", nn.String()), w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := d.client.Delete(context.TODO(), found); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
