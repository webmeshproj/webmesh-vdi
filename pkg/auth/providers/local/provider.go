// Package local contains an AuthProvider implementation backed by a passwd-like file
// stored in the secrets backend.
// This is primarily meant for testing, but could also be used in small setups.
package local

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthProvider implements an AuthProvider that uses a local secret similar
// to a passwd file to authenticate users and map them to roles. This is primarily
// intended for testing and ideally external auth providers would be supported.
type AuthProvider struct {
	common.AuthProvider

	// k8s client
	client client.Client
	// our cluster instance
	cluster *v1alpha1.VDICluster
	// the secrets engine where we store our passwd
	secrets *secrets.SecretEngine
}

// New returns a new AuthProvider.
func New() common.AuthProvider {
	return &AuthProvider{}
}

// Setup implements the AuthProvider interface and sets a local reference to the
// k8s client and vdi cluster.
func (a *AuthProvider) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	a.client = c
	a.cluster = cluster
	a.secrets = secrets.GetSecretEngine(cluster)
	return a.secrets.Setup(c, cluster)
}

// Close returns nil automatically as no cleanup is required.
func (a *AuthProvider) Close() error { return nil }
