package api

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceGetter struct {
	v1alpha1.ResourceGetter

	api *desktopAPI
}

func NewResourceGetter(d *desktopAPI) v1alpha1.ResourceGetter {
	return &ResourceGetter{api: d}
}

// Leaeving unimplemented. Only used by privilege escalation tests and checking
// usernames is not important.
func (r *ResourceGetter) GetUsers() ([]v1alpha1.VDIUser, error) {
	return []v1alpha1.VDIUser{}, nil
}

func (r *ResourceGetter) GetRoles() ([]v1alpha1.VDIRole, error) {
	roles, err := r.api.vdiCluster.GetRoles(r.api.client)
	if err != nil {
		apiLogger.Error(err, "Failed to list VDI roles")
		return nil, err
	}
	return roles, nil
}

func (r *ResourceGetter) GetTemplates() ([]v1alpha1.DesktopTemplate, error) {
	tmplList := &v1alpha1.DesktopTemplateList{}
	if err := r.api.client.List(context.TODO(), tmplList, client.InNamespace(metav1.NamespaceAll)); err != nil {
		apiLogger.Error(err, "Failed to list desktop templates")
		return nil, err
	}
	return tmplList.Items, nil
}
