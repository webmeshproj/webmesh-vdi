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
func (d *desktopAPI) PostLogout(w http.ResponseWriter, r *http.Request) {
	userSession := apiutil.GetRequestUserSession(r)
	if err := d.CleanupUserDesktops(userSession.User.GetName()); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	refreshToken, err := r.Cookie(RefreshTokenCookie)
	if err == nil {
		// Revoke the token and remove the cookie
		// Lookup will fetch and clear the token from the db.
		if _, err := d.lookupRefreshToken(refreshToken.Value); err != nil {
			apiLogger.Error(err, "Error while revoking refresh token, garbage may be left in the db")
		}
		// Set the cookie to an empty value
		http.SetCookie(w, &http.Cookie{
			Name:     RefreshTokenCookie,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
		})
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
