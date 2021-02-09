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

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceGetter satisfies the v1alpha1.ResourceGetter interface for retrieving
// available resources during a privilege check.
type ResourceGetter struct {
	v1.ResourceGetter
	// the underlying API object
	api *desktopAPI
}

// NewResourceGetter returns a new ResourceGetter
func NewResourceGetter(d *desktopAPI) v1.ResourceGetter {
	return &ResourceGetter{api: d}
}

// GetUsers is left unimplemented. Only used by privilege escalation tests
// and checking usernames is not important right now.
func (r *ResourceGetter) GetUsers() ([]v1.VDIUser, error) {
	return []v1.VDIUser{}, nil
}

// GetRoles returns a list of all the VDIRolse for this cluster.
func (r *ResourceGetter) GetRoles() ([]v1.VDIUserRole, error) {
	roles, err := r.api.vdiCluster.GetRoles(r.api.client)
	if err != nil {
		apiLogger.Error(err, "Failed to list VDI roles")
		return nil, err
	}
	userRoles := make([]v1.VDIUserRole, 0)
	for _, role := range roles {
		userRoles = append(userRoles, *role.ToUserRole())
	}
	return userRoles, nil
}

// GetTemplates returns a list of desktop templates for this cluster.
func (r *ResourceGetter) GetTemplates() ([]string, error) {
	tmplList := &v1alpha1.DesktopTemplateList{}
	if err := r.api.client.List(context.TODO(), tmplList, client.InNamespace(metav1.NamespaceAll)); err != nil {
		apiLogger.Error(err, "Failed to list desktop templates")
		return nil, err
	}
	tmplNames := make([]string, 0)
	for _, tmpl := range tmplList.Items {
		tmplNames = append(tmplNames, tmpl.GetName())
	}
	return tmplNames, nil
}
