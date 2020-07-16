package local

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LocalAuthProvider implements an AuthProvider that uses a local secret similar
// to a passwd file to authenticate users and map them to roles. This is primarily
// intended for testing and ideally external auth providers would be supported.
type LocalAuthProvider struct {
	v1alpha1.AuthProvider

	// k8s client
	client client.Client
	// our cluster instance
	cluster *v1alpha1.VDICluster
	// the secrets engine where we store our passwd
	secrets *secrets.SecretEngine
}

// New returns a new LocalAuthProvider.
func New() v1alpha1.AuthProvider {
	return &LocalAuthProvider{}
}

// Setup implements the AuthProvider interface and sets a local reference to the
// k8s client and vdi cluster.
func (a *LocalAuthProvider) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	a.client = c
	a.cluster = cluster
	a.secrets = secrets.GetSecretEngine(cluster)
	return a.secrets.Setup(c, cluster)
}

// Close returns nil automatically as no cleanup is required.
func (a *LocalAuthProvider) Close() error { return nil }
