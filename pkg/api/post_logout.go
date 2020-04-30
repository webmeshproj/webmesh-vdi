package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:route POST /api/logout Auth logout
// Ends the current user session.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) Logout(w http.ResponseWriter, r *http.Request) {
	userSession := apiutil.GetRequestUserSession(r)
	if err := d.CleanupUserDesktops(userSession.User.GetName()); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

func (d *desktopAPI) CleanupUserDesktops(username string) error {
	desktops := &v1alpha1.DesktopList{}
	if err := d.client.List(context.TODO(), desktops, client.InNamespace(metav1.NamespaceAll), d.vdiCluster.GetUserDesktopsSelector(username)); err != nil {
		return err
	}
	for _, item := range desktops.Items {
		if err := d.client.Delete(context.TODO(), &item); err != nil {
			return err
		}
	}
	return nil
}
