package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:route GET /api/config Miscellaneous getConfig
// Retrieves the current VDICluster configuration.
// responses:
//   200: configResponse
//   403: error
//   500: error
func (d *desktopAPI) GetConfig(w http.ResponseWriter, r *http.Request) {
	apiutil.WriteJSON(d.vdiCluster.Spec, w)
}

// Config response
// swagger:response configResponse
type swaggerConfigResponse struct {
	// in:body
	Body struct {
		v1alpha1.VDIClusterSpec
	}
}
