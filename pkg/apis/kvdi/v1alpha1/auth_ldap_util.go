package v1alpha1

import (
	"encoding/base64"
	"strings"
)

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
