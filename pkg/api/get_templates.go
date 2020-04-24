package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:route GET /api/templates Desktops getTemplates
// Retrieves available templates to boot desktops from.
// responses:
//   200: templatesResponse
//   400: error
//   403: error
//   500: error
func (d *desktopAPI) GetDesktopTemplates(w http.ResponseWriter, r *http.Request) {
	tmpls, err := d.getAllDesktopTemplates()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(tmpls.Items, w)
}

// getAllDesktopTemplates lists the DesktopTemplates registered in the api servers.
func (d *desktopAPI) getAllDesktopTemplates() (*v1alpha1.DesktopTemplateList, error) {
	tmplList := &v1alpha1.DesktopTemplateList{}
	return tmplList, d.client.List(context.TODO(), tmplList, client.InNamespace(metav1.NamespaceAll))
}

// Templates response
// swagger:response templatesResponse
type swaggerTemplatesResponse struct {
	// in:body
	Body []v1alpha1.DesktopTemplate
}
