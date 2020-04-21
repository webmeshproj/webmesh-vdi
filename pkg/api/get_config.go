package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

func (d *desktopAPI) GetConfig(w http.ResponseWriter, r *http.Request) {
	apiutil.WriteJSON(d.vdiCluster.Spec, w)
}
