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

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/user"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:route GET /api/templates Templates getTemplates
// Retrieves available templates to boot desktops from.
// responses:
//   200: templatesResponse
//   400: error
//   403: error
func (d *desktopAPI) GetDesktopTemplates(w http.ResponseWriter, r *http.Request) {
	sess := apiutil.GetRequestUserSession(r)
	tmpls, err := d.getAllDesktopTemplates()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(user.FilterTemplates(sess.User, tmpls.Items), w)
}

// getAllDesktopTemplates lists the DesktopTemplates registered in the api servers.
func (d *desktopAPI) getAllDesktopTemplates() (*v1alpha1.DesktopTemplateList, error) {
	tmplList := &v1alpha1.DesktopTemplateList{}
	return tmplList, d.client.List(context.TODO(), tmplList, client.InNamespace(metav1.NamespaceAll))
}

// swagger:operation GET /api/templates/{template} Templates getTemplate
// ---
// summary: Retrieve the specified DesktopTemplate.
// parameters:
// - name: template
//   in: path
//   description: The DesktopTemplate to retrieve details about
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/templateResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetDesktopTemplate(w http.ResponseWriter, r *http.Request) {
	tmplName := apiutil.GetTemplateFromRequest(r)
	nn := types.NamespacedName{Name: tmplName, Namespace: metav1.NamespaceAll}
	tmpl := &v1alpha1.DesktopTemplate{}
	if err := d.client.Get(context.TODO(), nn, tmpl); err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(tmpl, w)
}

// Templates response
// swagger:response templatesResponse
type swaggerTemplatesResponse struct {
	// in:body
	Body []v1alpha1.DesktopTemplate
}

// Templates response
// swagger:response templateResponse
type swaggerTemplateResponse struct {
	// in:body
	Body v1alpha1.DesktopTemplate
}
