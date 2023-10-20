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
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
)

// GetAdminSecret returns the name of the secret for storing the admin password.
func (c *VDICluster) GetAdminSecret() string {
	if c.Spec.Auth != nil && c.Spec.Auth.AdminSecret != "" {
		return c.Spec.Auth.AdminSecret
	}
	return fmt.Sprintf("%s-admin-secret", c.GetName())
}

// AnonymousAllowed returns true if anonymous users are allowed to interact with
// this cluster.
func (c *VDICluster) AnonymousAllowed() bool {
	if c.Spec.Auth != nil {
		return c.Spec.Auth.AllowAnonymous
	}
	return false
}

// IsUsingLocalAuth returns true if the cluster is using the local authentication
// driver. This function and the API should be refactored to just return true
// if no other options are defined.
func (c *VDICluster) IsUsingLocalAuth() bool {
	if c.Spec.Auth != nil {
		return c.Spec.Auth.LocalAuth != nil && !c.IsUsingLDAPAuth() && !c.IsUsingOIDCAuth() && !c.IsUsingWebmeshAuth()
	}
	return true
}

// IsUsingWebmeshAuth returns true if the cluster is using the webmesh authentication
// driver.
func (c *VDICluster) IsUsingWebmeshAuth() bool {
	if c.Spec.Auth != nil {
		return c.Spec.Auth.WebmeshAuth != nil && c.Spec.Auth.WebmeshAuth.MetadataURL != ""
	}
	return false
}

// AuthIsUsingSecretEngine returns true if the secrets for the configured auth
// backend are using the built-in secrets engine and not a separate kubernetes
// secret.
func (c *VDICluster) AuthIsUsingSecretEngine() bool {
	if c.Spec.Auth != nil {
		if c.Spec.Auth.LDAPAuth != nil {
			if c.Spec.Auth.LDAPAuth.BindCredentialsSecret != "" {
				return false
			}
		}
		if c.Spec.Auth.OIDCAuth != nil {
			if c.Spec.Auth.OIDCAuth.ClientCredentialsSecret != "" {
				return false
			}
		}
	}
	return true
}

// GetAuthK8sSecret returns the name of the k8s auth secret. For safety it returns
// the name of the app secret, however, the caller should only be using this function
// because they know they are not using the built-in secrets.
func (c *VDICluster) GetAuthK8sSecret() string {
	if c.Spec.Auth != nil {
		if c.Spec.Auth.LDAPAuth != nil && c.Spec.Auth.LDAPAuth.BindCredentialsSecret != "" {
			return c.Spec.Auth.LDAPAuth.BindCredentialsSecret
		}
		if c.Spec.Auth.OIDCAuth != nil && c.Spec.Auth.OIDCAuth.ClientCredentialsSecret != "" {
			return c.Spec.Auth.OIDCAuth.ClientCredentialsSecret
		}
	}
	return c.GetAppSecretsName()
}

// GetTokenDuration returns the duration for a new token to live. If the duration cannot be
// parsed, the default is returned
func (c *VDICluster) GetTokenDuration() time.Duration {
	if c.Spec.Auth != nil {
		if c.Spec.Auth.TokenDuration != "" {
			if duration, err := time.ParseDuration(c.Spec.Auth.TokenDuration); err == nil {
				return duration
			}
		}
	}
	return v1.DefaultSessionLength
}

// GetAdminRole returns an admin role for this VDICluster.
func (c *VDICluster) GetAdminRole() *rbacv1.VDIRole {
	var annotations map[string]string
	if c.IsUsingLDAPAuth() {
		annotations = map[string]string{
			v1.LDAPGroupRoleAnnotation: strings.Join(c.GetLDAPAdminGroups(), v1.AuthGroupSeparator),
		}
	} else if c.IsUsingOIDCAuth() {
		annotations = map[string]string{
			v1.OIDCGroupRoleAnnotation: strings.Join(c.GetOIDCAdminGroups(), v1.AuthGroupSeparator),
		}
	}
	return &rbacv1.VDIRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-admin", c.GetName()),
			Annotations: annotations,
			Labels: map[string]string{
				v1.RoleClusterRefLabel: c.GetName(),
			},
		},
		Rules: []rbacv1.Rule{
			{
				Verbs:            []rbacv1.Verb{rbacv1.VerbAll},
				Resources:        []rbacv1.Resource{rbacv1.ResourceAll},
				ResourcePatterns: []string{".*"},
				Namespaces:       []string{rbacv1.NamespaceAll},
			},
		},
	}
}
