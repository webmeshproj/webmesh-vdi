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

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	corev1 "k8s.io/api/core/v1"
)

// swagger:route GET /api/namespaces Miscellaneous getNamespaces
// Retrieves a list of namespaces the requesting user is allowed to provision desktops in.
// responses:
//   200: namespacesResponse
//   400: error
//   403: error
func (d *desktopAPI) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	sess := apiutil.GetRequestUserSession(r)
	namespaces, err := d.ListKubernetesNamespaces()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(sess.User.FilterNamespaces(namespaces), w)
}

// ListKubernetesNamespaces returns a string slice of all the namespaces
// in kubernetes.
func (d *desktopAPI) ListKubernetesNamespaces() ([]string, error) {
	nsList := &corev1.NamespaceList{}
	if err := d.client.List(context.TODO(), nsList); err != nil {
		return nil, err
	}
	nsNames := make([]string, 0)
	for _, ns := range nsList.Items {
		nsNames = append(nsNames, ns.GetName())
	}
	return nsNames, nil
}

// Namespaces response
// swagger:response namespacesResponse
type swaggerNamespacesResponse struct {
	// in:body
	Body []string
}
