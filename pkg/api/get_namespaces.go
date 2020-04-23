package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	corev1 "k8s.io/api/core/v1"
)

func (d *desktopAPI) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	sess := GetRequestUserSession(r)
	namespaces := sess.User.Namespaces()
	if namespaces == nil || len(namespaces) == 0 {
		var err error
		namespaces, err = d.ListKubernetesNamespaces()
		if err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
	}
	apiutil.WriteJSON(namespaces, w)
}

func (d *desktopAPI) ListKubernetesNamespaces() ([]string, error) {
	nsList := &corev1.NamespaceList{}
	if err := d.client.List(context.TODO(), nsList); err != nil {
		return nil, err
	}
	nsNames := make([]string, 0)
	for _, ns := range nsList.Items {
		nsNames = append(nsNames, ns.GetName())
	}
	return nsNames, nil
}
