package v1alpha1

import (
	"encoding/base64"

	oidc "github.com/coreos/go-oidc"
)

// IsUsingOIDCAuth returns true if the cluster is using the oidc authentication
// driver.
func (c *VDICluster) IsUsingOIDCAuth() bool {
	if c.Spec.Auth != nil {
		if c.Spec.Auth.OIDCAuth != nil && !c.Spec.Auth.OIDCAuth.IsUndefined() {
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

// GetOIDCClientIDKey returns the key in the secret where the client ID can be retrieved.
func (c *VDICluster) GetOIDCClientIDKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.ClientIDKey != "" {
			return c.Spec.Auth.OIDCAuth.ClientIDKey
		}
	}
	return "oidc-clientid"
}

// GetOIDCClientSecretKey returns the key in the secret where client secret can be retrieved.
func (c *VDICluster) GetOIDCClientSecretKey() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.ClientSecretKey != "" {
			return c.Spec.Auth.OIDCAuth.ClientSecretKey
		}
	}
	return "oidc-clientsecret"
}

// GetOIDCScopes returns the list of scopes to request from the OpenID provider.
func (c *VDICluster) GetOIDCScopes() []string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.Scopes != nil {
			return c.Spec.Auth.OIDCAuth.Scopes
		}
	}
	return []string{oidc.ScopeOpenID, "email", "profile", "groups"}
}

// GetOIDCGroupScope returns the scope to use for matching a user's groups to VDI roles.
func (c *VDICluster) GetOIDCGroupScope() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.GroupScope != "" {
			return c.Spec.Auth.OIDCAuth.GroupScope
		}
	}
	return "groups"
}

// GetOIDCAdminGroups returns the values in the groups claim that will map to administrator access.
func (c *VDICluster) GetOIDCAdminGroups() []string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.AdminGroups
	}
	return []string{}
}

// GetOIDCInsecureSkipVerify returns whether or not to verify the TLS certificate of the OIDC provider.
func (c *VDICluster) GetOIDCInsecureSkipVerify() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.TLSInsecureSkipVerify
	}
	return false
}

// GetOIDCCA returns the CA certificate to use when verifying the OIDC provider certificate. The
// value is base64 decoded and returned to the caller.
func (c *VDICluster) GetOIDCCA() ([]byte, error) {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		if c.Spec.Auth.OIDCAuth.TLSCACert != "" {
			return base64.StdEncoding.DecodeString(c.Spec.Auth.OIDCAuth.TLSCACert)
		}
	}
	return nil, nil
}

// GetOIDCRedirectURL returns the URL that the OIDC provider should redirect to after a successful
// authentication.
func (c *VDICluster) GetOIDCRedirectURL() string {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.RedirectURL
	}
	return ""
}

// AllowNonGroupedReadOnly returns true if non-grouped users from the OpenID provider should
// be allowed read-only access to kVDI.
func (c *VDICluster) AllowNonGroupedReadOnly() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.AllowNonGroupedReadOnly
	}
	return false
}

// PreserveOIDCTokens returns whether OIDC tokens should be preserved and stored in the kvdi claims
// for the user.
func (c *VDICluster) PreserveOIDCTokens() bool {
	if c.Spec.Auth != nil && c.Spec.Auth.OIDCAuth != nil {
		return c.Spec.Auth.OIDCAuth.PreserveTokens
	}
	return false
}
