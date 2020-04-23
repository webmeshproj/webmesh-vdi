package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *desktopAPI) Logout(w http.ResponseWriter, r *http.Request) {
	userSession := GetRequestUserSession(r)
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	if err := sess.DeleteUserSession(userSession); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := d.CleanupUserDesktops(userSession.User.Name); err != nil {
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
