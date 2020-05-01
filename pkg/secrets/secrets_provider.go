package secrets

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets/k8secret"
)

func GetProvider(cluster *v1alpha1.VDICluster) v1alpha1.SecretsProvider {
	return k8secret.New()
}
