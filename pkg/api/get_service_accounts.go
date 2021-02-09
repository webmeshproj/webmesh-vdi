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
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

// swagger:operation GET /api/serviceaccounts/{namespace} Miscellaneous getServiceAccounts
// ---
// summary: Retrieve the service accounts in the given namespace that the user is allowed to use
// parameters:
// - name: namespace
//   in: path
//   description: The namespace to list service accounts in
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/serviceAccountsResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetServiceAccounts(w http.ResponseWriter, r *http.Request) {
	namespace := apiutil.GetNamespaceFromRequest(r)
	sess := apiutil.GetRequestUserSession(r)
	serviceAccounts, err := d.ListServiceAccounts(namespace)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(sess.User.FilterServiceAccounts(serviceAccounts, namespace), w)
}

// ListServiceAccounts returns a string slice of all the service accounts
// in a given namespace.
func (d *desktopAPI) ListServiceAccounts(ns string) ([]string, error) {
	saList := &corev1.ServiceAccountList{}
	if err := d.client.List(context.TODO(), saList, client.InNamespace(ns)); err != nil {
		return nil, err
	}
	saNames := make([]string, 0)
	for _, sa := range saList.Items {
		saNames = append(saNames, sa.GetName())
	}
	return saNames, nil
}

// Service Accounts Response
// swagger:response serviceAccountsResponse
type swaggerServiceAccountsResponse struct {
	// in:body
	Body []string
}
