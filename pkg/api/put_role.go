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
	"fmt"
	"net/http"

	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"

	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation PUT /api/roles/{role} Roles putRoleRequest
// ---
// summary: Update the specified role.
// description: All properties will be overwritten with those provided in the payload, even if undefined.
// parameters:
//   - name: role
//     in: path
//     description: The role to update
//     type: string
//     required: true
//   - in: body
//     name: roleDetails
//     description: The role details to update.
//     schema:
//     "$ref": "#/definitions/UpdateRoleRequest"
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/boolResponse"
//	"400":
//	  "$ref": "#/responses/error"
//	"403":
//	  "$ref": "#/responses/error"
//	"404":
//	  "$ref": "#/responses/error"
func (d *desktopAPI) UpdateRole(w http.ResponseWriter, r *http.Request) {
	role := apiutil.GetRoleFromRequest(r)
	nn := ktypes.NamespacedName{Name: role, Namespace: metav1.NamespaceAll}
	vdiRole := &rbacv1.VDIRole{}
	if err := d.client.Get(context.TODO(), nn, vdiRole); err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(fmt.Errorf("The role '%s' doesn't exist", role), w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	params := apiutil.GetRequestObject(r).(*types.UpdateRoleRequest)
	if params == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	vdiRole.Annotations = params.GetAnnotations()
	vdiRole.Rules = params.GetRules()
	if err := d.client.Update(context.TODO(), vdiRole); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// Request containing updates to a role
// swagger:parameters putRoleRequest
type swaggerUpdateRoleRequest struct {
	// in:body
	Body types.UpdateRoleRequest
}
