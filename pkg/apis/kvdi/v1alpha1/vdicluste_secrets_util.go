package v1alpha1

import "strings"

const (
	SecretsBackendK8s   = "k8s"
	SecretsBackendVault = "vault"
)

func (c *VDICluster) GetSecretsBackend() string {
	if c.Spec.Secrets != nil {
		if c.Spec.Secrets.Vault != nil {
			return SecretsBackendVault
		}
	}
	return SecretsBackendK8s
}

func (v *VaultConfig) GetAuthRole() string {
	if v.AuthRole != "" {
		return v.AuthRole
	}
	return "kvdi"
}

func (v *VaultConfig) GetSecretsPath() string {
	if v.SecretsPath != "" {
		return strings.TrimSuffix(v.SecretsPath, "/")
	}
	return "kvdi"
}
