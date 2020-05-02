package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:route POST /api/config/reload Miscellaneous postReload
// Reloads the server configuration.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) PostReloadConfig(w http.ResponseWriter, r *http.Request) {
	cluster := &v1alpha1.VDICluster{}
	// retrieve the latest cluster configuration
	if err := d.client.Get(context.TODO(), d.vdiCluster.NamespacedName(), cluster); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	d.vdiCluster = cluster
	// reload the cluster configuration into the secrets backend
	if err := d.secrets.Setup(d.client, cluster); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
