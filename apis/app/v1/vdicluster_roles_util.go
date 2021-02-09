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

package v1

import (
	"context"
	"fmt"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	rbacv1 "github.com/tinyzimmer/kvdi/apis/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetRoles returns a list of all the VDIRoles that apply to this cluster instance.
func (c *VDICluster) GetRoles(cl client.Client) ([]rbacv1.VDIRole, error) {
	roleList := &rbacv1.VDIRoleList{}
	return roleList.Items, cl.List(
		context.TODO(),
		roleList,
		client.InNamespace(metav1.NamespaceAll),
		client.MatchingLabels{v1.RoleClusterRefLabel: c.GetName()},
	)
}

// GetLaunchTemplatesRole returns a launch-templates role for a cluster. A role like this
// is created for every cluster for convenience. It is the default role applied to anonymous
// users, and for non-grouped OIDC users.
func (c *VDICluster) GetLaunchTemplatesRole() *rbacv1.VDIRole {
	role := &rbacv1.VDIRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-launch-templates", c.GetName()),
			OwnerReferences: c.OwnerReferences(),
			Labels: map[string]string{
				v1.RoleClusterRefLabel: c.GetName(),
			},
		},
	}
	if c.Spec.Auth != nil && c.Spec.Auth.DefaultRoleRules != nil {
		role.Rules = c.Spec.Auth.DefaultRoleRules
	} else {
		role.Rules = []rbacv1.Rule{
			{
				Verbs:            []rbacv1.Verb{rbacv1.VerbRead, rbacv1.VerbUse, rbacv1.VerbLaunch},
				Resources:        []rbacv1.Resource{rbacv1.ResourceTemplates},
				ResourcePatterns: []string{".*"},
				Namespaces:       []string{c.GetCoreNamespace()},
			},
		}
	}
	return role
}
