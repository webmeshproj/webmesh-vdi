package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDesktopTemplates mirrors an API call to list all the available desktop
// templates back to the caller.
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
