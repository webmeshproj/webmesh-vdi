package api

import (
	"net/http"
	"regexp"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// PostRoleRequest requests updates to an existing role.
type PutRoleRequest struct {
	Grants           grants.RoleGrant `json:"grants"`
	Namespaces       []string         `json:"namespaces"`
	TemplatePatterns []string         `json:"templatePatterns"`
}

func (p *PutRoleRequest) Validate() error {
	for _, x := range p.TemplatePatterns {
		if _, err := regexp.Compile(x); err != nil {
			return err
		}
	}
	return nil
}

func newRoleFromPutRequest(name string, req *PutRoleRequest) *types.Role {
	return &types.Role{
		Name:             name,
		Grants:           req.Grants,
		Namespaces:       req.Namespaces,
		TemplatePatterns: req.TemplatePatterns,
	}
}

// swagger:operation PUT /api/roles/{role} Roles putRoleRequest
// ---
// summary: Update the specified role.
// description: All properties will be overwritten with those provided in the payload, even if undefined.
// parameters:
// - name: role
//   in: path
//   description: The role to update
//   type: string
//   required: true
// - in: body
//   name: roleDetails
//   description: The role details to update.
//   schema:
//     "$ref": "#/definitions/PutRoleRequest"
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "500":
//     "$ref": "#/responses/error"
func (d *desktopAPI) UpdateRole(w http.ResponseWriter, r *http.Request) {
	req := GetRequestObject(r).(*PutRoleRequest)
	if err := req.Validate(); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	role := newRoleFromPutRequest(getRoleFromRequest(r), req)
	sess, err := d.getDB()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	if err := sess.UpdateRole(role); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// Request containing updates to a role
// swagger:parameters putRoleRequest
type swaggerUpdateRoleRequest struct {
	// in:body
	Body PutRoleRequest
}
