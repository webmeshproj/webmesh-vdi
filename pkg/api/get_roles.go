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
	"fmt"
	"net/http"

	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
)

// swagger:route GET /api/roles Roles getRoles
// Retrieves a list of the authorization roles in kVDI.
// responses:
//
//	200: rolesResponse
//	400: error
//	403: error
func (d *desktopAPI) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := d.vdiCluster.GetRoles(d.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(roles, w)
}

// swagger:operation GET /api/roles/{role} Roles getRole
// ---
// summary: Retrieve the specified role.
// description: Details include the grants, namespaces, and template patterns for the role.
// parameters:
//   - name: role
//     in: path
//     description: The role to retrieve details about
//     type: string
//     required: true
//
// responses:
//
//	"200":
//	  "$ref": "#/responses/roleResponse"
//	"400":
//	  "$ref": "#/responses/error"
//	"403":
//	  "$ref": "#/responses/error"
//	"404":
//	  "$ref": "#/responses/error"
func (d *desktopAPI) GetRole(w http.ResponseWriter, r *http.Request) {
	roles, err := d.vdiCluster.GetRoles(d.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	roleName := apiutil.GetRoleFromRequest(r)
	for _, role := range roles {
		if role.GetName() == roleName {
			apiutil.WriteJSON(role, w)
			return
		}
	}
	apiutil.ReturnAPINotFound(fmt.Errorf("No role with the name '%s' found", roleName), w)
}

// A list of roles
// swagger:response rolesResponse
type swaggerRolesResponse struct {
	// in:body
	Body []rbacv1.VDIRole
}

// A single role
// swagger:response roleResponse
type swaggerRoleResponse struct {
	// in:body
	Body rbacv1.VDIRole
}
