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

// swagger:route GET /api/templates Templates getTemplates
// Retrieves available templates to boot desktops from.
// responses:
//   200: templatesResponse
//   400: error
//   403: error
func (d *desktopAPI) GetDesktopTemplates(w http.ResponseWriter, r *http.Request) {
	sess := apiutil.GetRequestUserSession(r)
	tmpls, err := d.getAllDesktopTemplates()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(sess.User.FilterTemplates(tmpls.Items), w)
}

// getAllDesktopTemplates lists the DesktopTemplates registered in the api servers.
func (d *desktopAPI) getAllDesktopTemplates() (*v1alpha1.DesktopTemplateList, error) {
	tmplList := &v1alpha1.DesktopTemplateList{}
	return tmplList, d.client.List(context.TODO(), tmplList, client.InNamespace(metav1.NamespaceAll))
}

// swagger:operation GET /api/templates/{template} Templates getTemplate
// ---
// summary: Retrieve the specified DesktopTemplate.
// parameters:
// - name: template
//   in: path
//   description: The DesktopTemplate to retrieve details about
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/templateResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopTemplate(w http.ResponseWriter, r *http.Request) {
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
	apiutil.WriteJSON(tmpl, w)
}

// Templates response
// swagger:response templatesResponse
type swaggerTemplatesResponse struct {
	// in:body
	Body []v1alpha1.DesktopTemplate
}

// Templates response
// swagger:response templateResponse
type swaggerTemplateResponse struct {
	// in:body
	Body v1alpha1.DesktopTemplate
}
