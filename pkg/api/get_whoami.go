package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:route GET /api/whoami Miscellaneous whoAmI
// Retrieves information about the current user session.
// responses:
//   200: userResponse
//   403: error
//   500: error
func (d *desktopAPI) GetWhoAmI(w http.ResponseWriter, r *http.Request) {
	// retrieve the user session from the request
	session := apiutil.GetRequestUserSession(r)
	// retrieve all desktops for this user and populate the Sessions field
	desktops := &v1alpha1.DesktopList{}
	if err := d.client.List(context.TODO(), desktops, client.InNamespace(metav1.NamespaceAll), d.vdiCluster.GetUserDesktopsSelector(session.User.Name)); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	// If the user has any active desktops, append them to the response
	if len(desktops.Items) > 0 {
		session.User.Sessions = make([]*v1.DesktopSession, len(desktops.Items))
		for idx, desktop := range desktops.Items {
			session.User.Sessions[idx] = &v1.DesktopSession{
				Name:      desktop.GetName(),
				Namespace: desktop.GetNamespace(),
				User:      desktop.GetUser(),
			}
		}
	}
	apiutil.WriteJSON(session.User, w)
}
