package v1alpha1

import (
	"fmt"
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
		if c.Spec.Auth.LocalAuth != nil {
			return true
		}
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
