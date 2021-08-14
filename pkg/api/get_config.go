/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package api

import (
	"net/http"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
)

// swagger:route GET /api/config Miscellaneous getConfig
// Retrieves the current VDICluster configuration.
// responses:
//   200: configResponse
//   400: error
//   403: error
func (d *desktopAPI) GetConfig(w http.ResponseWriter, r *http.Request) {
	apiutil.WriteJSON(d.vdiCluster.Spec, w)
}

// Config response
// swagger:response configResponse
type swaggerConfigResponse struct {
	// in:body
	Body struct {
		appv1.VDIClusterSpec
	}
}
