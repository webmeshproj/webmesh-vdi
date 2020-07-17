package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
//     "$ref": "#/definitions/UpdateRoleRequest"
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) UpdateRole(w http.ResponseWriter, r *http.Request) {
	role := apiutil.GetRoleFromRequest(r)
	nn := types.NamespacedName{Name: role, Namespace: metav1.NamespaceAll}
	vdiRole := &v1alpha1.VDIRole{}
	if err := d.client.Get(context.TODO(), nn, vdiRole); err != nil {
		if client.IgnoreNotFound(err) == nil {
			apiutil.ReturnAPINotFound(fmt.Errorf("The role '%s' doesn't exist", role), w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}
	params := apiutil.GetRequestObject(r).(*v1alpha1.UpdateRoleRequest)
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
	Body v1alpha1.UpdateRoleRequest
}
