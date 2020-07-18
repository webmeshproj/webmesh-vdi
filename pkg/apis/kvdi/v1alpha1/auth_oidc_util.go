package v1alpha1

import (
	"encoding/base64"

	oidc "github.com/coreos/go-oidc"
)

// IsUsingOIDCAuth returns true if the cluster is using the oidc authentication
// driver.
func (c *VDICluster) IsUsingOIDCAuth() bool {
	if c.Spec.Auth != nil {
		if c.Spec.Auth.OIDCAuth != nil {
			return true
		}
	}
	return false
}

// GetOIDCIssuerURL returns the OIDC issuer URL or a blank string (which will
// throw an error when used).
func (c *VDICluster) GetOIDCIssuerURL() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.IssuerURL
	}
	return ""
}

func (c *VDICluster) GetOIDCClientIDKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.ClientIDKey != "" {
			return c.Spec.Auth.OIDCAuth.ClientIDKey
		}
	}
	return "oidc-clientid"
}

func (c *VDICluster) GetOIDCClientSecretKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.ClientSecretKey != "" {
			return c.Spec.Auth.OIDCAuth.ClientSecretKey
		}
	}
	return "oidc-clientsecret"
}

func (c *VDICluster) GetOIDCScopes() []string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.Scopes != nil {
			return c.Spec.Auth.OIDCAuth.Scopes
		}
	}
	return []string{oidc.ScopeOpenID, "email", "profile", "groups"}
}

func (c *VDICluster) GetOIDCGroupScope() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.GroupScope != "" {
			return c.Spec.Auth.OIDCAuth.GroupScope
		}
	}
	return "groups"
}

func (c *VDICluster) GetOIDCAdminGroups() []string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.AdminGroups
	}
	return []string{}
}

func (c *VDICluster) GetOIDCInsecureSkipVerify() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.TLSInsecureSkipVerify
	}
	return false
}

func (c *VDICluster) GetOIDCCA() ([]byte, error) {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.TLSCACert != "" {
			return base64.StdEncoding.DecodeString(c.Spec.Auth.OIDCAuth.TLSCACert)
		}
	}
	return nil, nil
}

func (c *VDICluster) GetOIDCRedirectURL() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.RedirectURL
	}
	return ""
}

func (c *VDICluster) AllowNonGroupedReadOnly() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.AllowNonGroupedReadOnly
	}
	return false
}
