package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetRoles returns a list of all the VDIRoles that apply to this cluster instance.
func (v *VDICluster) GetRoles(c client.Client) ([]VDIRole, error) {
	roleList := &VDIRoleList{}
	return roleList.Items, c.List(
		context.TODO(),
		roleList,
		client.InNamespace(metav1.NamespaceAll),
		client.MatchingLabels{RoleClusterRefLabel: v.GetName()},
	)
}
