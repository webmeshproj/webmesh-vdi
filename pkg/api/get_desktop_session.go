package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation GET /api/sessions/{namespace}/{name} Desktops getSession
// ---
// summary: Retrieve the status of the requested desktop session.
// description: Details include the podPhase and CRD status.
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
//     "$ref": "#/responses/getSessionResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopSessionStatus(w http.ResponseWriter, r *http.Request) {
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
	res := make(map[string]interface{})
	res["running"] = found.Status.Running
	res["podPhase"] = found.Status.PodPhase
	apiutil.WriteJSON(res, w)
}

// Session status response
// swagger:response getSessionResponse
type swaggerGetSessionResponse struct {
	// in:body
	Body map[string]interface{}
}
