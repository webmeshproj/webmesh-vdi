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

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"

	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Request containing a new user
// swagger:parameters postRoleRequest
type swaggerCreateRoleRequest struct {
	// in:body
	Body types.CreateRoleRequest
}

// swagger:route POST /api/roles Roles postRoleRequest
// Create a new role in kVDI.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) CreateRole(w http.ResponseWriter, r *http.Request) {
	req := apiutil.GetRequestObject(r).(*types.CreateRoleRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	role := d.newRoleFromRequest(req)
	if err := d.client.Create(context.TODO(), role); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

func (d *desktopAPI) newRoleFromRequest(req *types.CreateRoleRequest) *rbacv1.VDIRole {
	return &rbacv1.VDIRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.GetName(),
			Annotations: req.GetAnnotations(),
			Labels: map[string]string{
				v1.RoleClusterRefLabel: d.vdiCluster.GetName(),
			},
		},
		Rules: req.GetRules(),
	}
}
