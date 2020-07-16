package auth

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/ldap"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/local"
)

// GetAuthProvider returns the authentication provider for the given VDICluster.
// Currently only local-auth is supported.
func GetAuthProvider(cluster *v1alpha1.VDICluster) v1alpha1.AuthProvider {
	if cluster.IsUsingLDAPAuth() {
		return ldap.New()
	}
	return local.New()
}
