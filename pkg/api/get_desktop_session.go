package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// GetSessionStatus returns to the caller whether the instance is running and
// resolveable inside the cluster.
func (d *desktopAPI) GetDesktopSessionStatus(w http.ResponseWriter, r *http.Request) {
	nn := getNamespacedNameFromRequest(r)
	found := &v1alpha1.Desktop{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	res := make(map[string]interface{})
	res["running"] = found.Status.Running
	res["podPhase"] = found.Status.PodPhase
	apiutil.WriteJSON(res, w)
}
