// Package auth contains the methods for retrieving the AuthProvider for a
// given cluster.
package auth

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/ldap"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/local"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/oidc"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
)

// GetAuthProvider returns the authentication provider for the given VDICluster. The secret engine passed
// to the provider is assumed to already be setup.
func GetAuthProvider(cluster *v1alpha1.VDICluster, s *secrets.SecretEngine) common.AuthProvider {
	if cluster.IsUsingLDAPAuth() {
		return ldap.New(s)
	}
	if cluster.IsUsingOIDCAuth() {
		return oidc.New(s)
	}
	return local.New(s)
}
