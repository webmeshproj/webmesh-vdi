package api

import (
	"context"
	"net"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/types"
)

// getNamespacedNameFromRequest returns the namespaced name of the Desktop instance
// for the given request.
func getNamespacedNameFromRequest(r *http.Request) types.NamespacedName {
	vars := mux.Vars(r)
	return types.NamespacedName{Name: vars["name"], Namespace: vars["namespace"]}
}

// GetSessionStatus returns to the caller whether the instance is running and
// resolveable inside the cluster.
func (d *desktopAPI) GetSessionStatus(w http.ResponseWriter, r *http.Request) {
	nn := getNamespacedNameFromRequest(r)
	found := &v1alpha1.Desktop{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	res := make(map[string]interface{})
	res["running"] = found.Status.Running
	if _, err := net.LookupHost(util.DesktopShortURL(found)); err == nil {
		res["resolvable"] = true
	} else {
		res["resolvable"] = false
	}
	apiutil.WriteJSON(res, w)
}
