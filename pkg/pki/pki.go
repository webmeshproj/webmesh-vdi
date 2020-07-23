package pki

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager provides certificate generation, signing, and storage for
// mTLS communication in a VDICluster.
type Manager struct {
	cluster *v1alpha1.VDICluster
	client  client.Client
	secrets *secrets.SecretEngine
}

// New returns a new PKI manager for the provided VDICluster.
func New(c client.Client, cluster *v1alpha1.VDICluster, s *secrets.SecretEngine) *Manager {
	return &Manager{
		cluster: cluster,
		client:  c,
		secrets: s,
	}
}
