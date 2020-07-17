package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Request containing a new user
// swagger:parameters postRoleRequest
type swaggerCreateRoleRequest struct {
	// in:body
	Body v1alpha1.CreateRoleRequest
}

// swagger:route POST /api/roles Roles postRoleRequest
// Create a new role in kVDI.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) CreateRole(w http.ResponseWriter, r *http.Request) {
	req := apiutil.GetRequestObject(r).(*v1alpha1.CreateRoleRequest)
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

func (d *desktopAPI) newRoleFromRequest(req *v1alpha1.CreateRoleRequest) *v1alpha1.VDIRole {
	return &v1alpha1.VDIRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.GetName(),
			Annotations: req.GetAnnotations(),
			Labels: map[string]string{
				v1alpha1.RoleClusterRefLabel: d.vdiCluster.GetName(),
			},
		},
		Rules: req.GetRules(),
	}
}
