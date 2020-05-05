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

// swagger:operation PUT /api/templates/{template} Templates putTemplateRequest
// ---
// summary: Update the specified DesktopTemplate.
// description: Only attributes defined in the payload will be applied.
// parameters:
// - name: template
//   in: path
//   description: The DesktopTemplate to update
//   type: string
//   required: true
// - in: body
//   name: templateDetails
//   description: The manifest to merge with the existing template.
//   schema:
//     "$ref": "#/definitions/DesktopTemplate"
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) PutDesktopTemplate(w http.ResponseWriter, r *http.Request) {
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
	// This will replace fields in the existing object with any provided in the
	// payload
	if err := apiutil.UnmarshalRequest(r, tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	if err := d.client.Update(context.TODO(), tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteOK(w)
}

// Request containing updates to a template
// swagger:parameters putTemplateRequest
type swaggerUpdateTemplateRequest struct {
	// in:body
	Body v1alpha1.DesktopTemplate
}
