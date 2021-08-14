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
	"context"
	"errors"
	"net/http"

	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
)

// swagger:route POST /api/templates Templates postTemplateRequest
// Create a new DesktopTemplate in kVDI.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) PostDesktopTemplates(w http.ResponseWriter, r *http.Request) {
	tmpl := apiutil.GetRequestObject(r).(*desktopsv1.Template)
	if tmpl == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	if err := d.client.Create(context.TODO(), tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// Request containing a new user
// swagger:parameters postTemplateRequest
type swaggerCreateTemplateRequest struct {
	// in:body
	Body desktopsv1.Template
}
