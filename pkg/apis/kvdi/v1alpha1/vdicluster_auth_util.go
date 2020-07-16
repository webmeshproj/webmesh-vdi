package v1alpha1

import (
	"encoding/base64"
	"fmt"
	"strings"
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

// IsUsingLDAPAuth returns true if the cluster is using the ldap authentication
// driver.
func (c *VDICluster) IsUsingLDAPAuth() bool {
	if c.Spec.Auth != nil {
		if c.Spec.Auth.LDAPAuth != nil {
			return true
		}
	}
	return false
}

func (c *VDICluster) GetLDAPURL() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.URL
	}
	return ""
}

// IsUsingLDAPOverTLS returns true if the configured LDAP server is using TLS.
func (c *VDICluster) IsUsingLDAPOverTLS() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.URL != "" {
			if strings.HasPrefix(c.Spec.Auth.LDAPAuth.URL, "ldaps") {
				return true
			}
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
	}
	return true
}

// GetAuthK8sSecret returns the name of the k8s auth secret. For safety it returns
// the name of the app secret, however, the caller should only be using this function
// because they know they are not using the built-in secrets.
func (c *VDICluster) GetAuthK8sSecret() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil && c.Spec.Auth.LDAPAuth.BindCredentialsSecret != "" {
		return c.Spec.Auth.LDAPAuth.BindCredentialsSecret
	}
	return c.GetAppSecretsName()
}

func (c *VDICluster) GetLDAPUserDNKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.BindUserDNSecretKey != "" {
			return c.Spec.Auth.LDAPAuth.BindUserDNSecretKey
		}
	}
	return "ldap-userdn"
}

func (c *VDICluster) GetLDAPPasswordKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.BindPasswordSecretKey != "" {
			return c.Spec.Auth.LDAPAuth.BindPasswordSecretKey
		}
	}
	return "ldap-password"
}

func (c *VDICluster) GetLDAPInsecureSkipVerify() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.TLSInsecureSkipVerify
	}
	return false
}

func (c *VDICluster) GetLDAPCA() ([]byte, error) {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.TLSCACert != "" {
			return base64.StdEncoding.DecodeString(c.Spec.Auth.LDAPAuth.TLSCACert)
		}
	}
	return nil, nil
}

func (c *VDICluster) GetLDAPSearchBase() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.UserSearchBase
	}
	return ""
}

func (c *VDICluster) GetLDAPAdminGroups() []string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.AdminGroups
	}
	return []string{}
}
