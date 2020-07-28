// Package auth contains the methods for retrieving the AuthProvider for a
// given cluster.
package auth

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/tinyzimmer/kvdi/pkg/auth/common"

	"github.com/tinyzimmer/kvdi/pkg/auth/providers/ldap"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/local"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/oidc"
)

// GetAuthProvider returns the authentication provider for the given VDICluster.
// Currently only local-auth is supported.
func GetAuthProvider(cluster *v1alpha1.VDICluster) common.AuthProvider {
	if cluster.IsUsingLDAPAuth() {
		return ldap.New()
	}
	if cluster.IsUsingOIDCAuth() {
		return oidc.New()
	}
	return local.New()
}
