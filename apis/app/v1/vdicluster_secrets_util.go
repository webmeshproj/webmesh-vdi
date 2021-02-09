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
		if c.Spec.Secrets.Vault != nil && !c.Spec.Secrets.Vault.IsUndefined() {
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
