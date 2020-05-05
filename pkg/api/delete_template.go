package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation DELETE /api/templates/{template} Templates deleteTemplateRequest
// ---
// summary: Delete the specified DesktopTemplate.
// parameters:
// - name: template
//   in: path
//   description: The DesktopTemplate to delete
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
func (d *desktopAPI) DeleteDesktopTemplate(w http.ResponseWriter, r *http.Request) {
	tmplName := apiutil.GetTemplateFromRequest(r)
	nn := types.NamespacedName{Name: tmplName, Namespace: metav1.NamespaceAll}
	tmpl := &v1alpha1.DesktopTemplate{}
	if err := d.client.Get(context.TODO(), nn, tmpl); err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := d.client.Delete(context.TODO(), tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
	}
	apiutil.WriteOK(w)
}
