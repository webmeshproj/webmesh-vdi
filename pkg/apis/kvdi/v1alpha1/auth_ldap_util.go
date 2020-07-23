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

// GetLDAPURL returns the full URL to the configured LDAP server.
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

// GetLDAPUserDNKey returns the key in the secret where the bind DN can be retrieved.
func (c *VDICluster) GetLDAPUserDNKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.BindUserDNSecretKey != "" {
			return c.Spec.Auth.LDAPAuth.BindUserDNSecretKey
		}
	}
	return "ldap-userdn"
}

// GetLDAPPasswordKey returns the key in the secret where the bind password can be retrieved.
func (c *VDICluster) GetLDAPPasswordKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.BindPasswordSecretKey != "" {
			return c.Spec.Auth.LDAPAuth.BindPasswordSecretKey
		}
	}
	return "ldap-password"
}

// GetLDAPInsecureSkipVerify returns whether TLS certificate verification should be performed on the LDAPS connection.
func (c *VDICluster) GetLDAPInsecureSkipVerify() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.TLSInsecureSkipVerify
	}
	return false
}

// GetLDAPCA returns the CA certificate to use when verifying the LDAPS server certificate.
// The configured result is base64 decoded and sent back to the caller.
func (c *VDICluster) GetLDAPCA() ([]byte, error) {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		if c.Spec.Auth.LDAPAuth.TLSCACert != "" {
			return base64.StdEncoding.DecodeString(c.Spec.Auth.LDAPAuth.TLSCACert)
		}
	}
	return nil, nil
}

// GetLDAPSearchBase returns the base DN to use when querying users from LDAP.
func (c *VDICluster) GetLDAPSearchBase() string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.UserSearchBase
	}
	return ""
}

// GetLDAPAdminGroups returns the list of groups in LDAP that should be bound to the kvdi-admin
// role.
func (c *VDICluster) GetLDAPAdminGroups() []string {
	if c.Spec.Auth != nil && c.Spec.Auth.LDAPAuth != nil {
		return c.Spec.Auth.LDAPAuth.AdminGroups
	}
	return []string{}
}