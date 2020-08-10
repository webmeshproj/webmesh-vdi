package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:operation DELETE /api/roles/{role} Roles deleteRoleRequest
// ---
// summary: Delete the specified role.
// parameters:
// - name: role
//   in: path
//   description: The role to delete
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/boolResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) DeleteRole(w http.ResponseWriter, r *http.Request) {
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
	if err := d.client.Delete(context.TODO(), vdiRole); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
