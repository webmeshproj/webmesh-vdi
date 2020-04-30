package api

import (
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:route GET /api/roles Roles getRoles
// Retrieves a list of the authorization roles in kVDI.
// responses:
//   200: rolesResponse
//   400: error
//   403: error
func (d *desktopAPI) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := d.vdiCluster.GetRoles(d.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
	}
	apiutil.WriteJSON(roles, w)
}

// swagger:operation GET /api/roles/{role} Roles getRole
// ---
// summary: Retrieve the specified role.
// description: Details include the grants, namespaces, and template patterns for the role.
// parameters:
// - name: role
//   in: path
//   description: The role to retrieve details about
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/roleResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetRole(w http.ResponseWriter, r *http.Request) {
	roles, err := d.vdiCluster.GetRoles(d.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
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
	Body []v1alpha1.VDIRole
}

// A single role
// swagger:response roleResponse
type swaggerRoleResponse struct {
	// in:body
	Body v1alpha1.VDIRole
}
