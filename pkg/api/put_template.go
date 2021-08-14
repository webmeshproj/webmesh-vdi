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
	"net/http"

	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	"github.com/kvdi/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation PUT /api/templates/{template} Templates putTemplateRequest
// ---
// summary: Update the specified DesktopTemplate.
// description: Only attributes defined in the payload will be applied.
// parameters:
// - name: template
//   in: path
//   description: The DesktopTemplate to update
//   type: string
//   required: true
// - in: body
//   name: templateDetails
//   description: The manifest to merge with the existing template.
//   schema:
//     "$ref": "#/definitions/DesktopTemplate"
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) PutDesktopTemplate(w http.ResponseWriter, r *http.Request) {
	tmplName := apiutil.GetTemplateFromRequest(r)
	nn := types.NamespacedName{Name: tmplName, Namespace: metav1.NamespaceAll}
	tmpl := &desktopsv1.Template{}
	if err := d.client.Get(context.TODO(), nn, tmpl); err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	// This will replace fields in the existing object with any provided in the
	// payload
	if err := apiutil.UnmarshalRequest(r, tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	if err := d.client.Update(context.TODO(), tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteOK(w)
}

// Request containing updates to a template
// swagger:parameters putTemplateRequest
type swaggerUpdateTemplateRequest struct {
	// in:body
	Body desktopsv1.Template
}
