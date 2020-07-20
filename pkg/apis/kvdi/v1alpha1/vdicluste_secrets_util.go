package v1alpha1

import "strings"

const (
	// SecretsBackendK8s represents using a kubernetes secret for secret storage.
	SecretsBackendK8s = "k8s"
	// SecretsBackendVault represents using vault for secret storage.
	SecretsBackendVault = "vault"
)

// GetSecretsBackend returns the type of secrets backend this VDICluster is using.
func (c *VDICluster) GetSecretsBackend() string {
	if c.Spec.Secrets != nil {
		if c.Spec.Secrets.Vault != nil {
			return SecretsBackendVault
		}
	}
	return SecretsBackendK8s
}

// GetAuthRole returns the auth role to use when connecting to a vault server.
func (v *VaultConfig) GetAuthRole() string {
	if v.AuthRole != "" {
		return v.AuthRole
	}
	return "kvdi"
}

// GetSecretsPath returns the path in vault to use for storing and retrieving secrets.
func (v *VaultConfig) GetSecretsPath() string {
	if v.SecretsPath != "" {
		return strings.TrimSuffix(v.SecretsPath, "/")
	}
	return "kvdi"
}
