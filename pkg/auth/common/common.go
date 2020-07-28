// Package common contains the core AuthProvider interface and utility functions
// to be used by the auth providers.
package common

import (
	"context"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetAuthSecrets is a helper function for retrieving multiple secrets required for
// authentication.
func GetAuthSecrets(c client.Client, cluster *v1alpha1.VDICluster, secrets *secrets.SecretEngine, keys ...string) (map[string]string, error) {
	results := make(map[string]string)
	if cluster.AuthIsUsingSecretEngine() {
		var res []byte
		var err error
		for _, key := range keys {
			if res, err = secrets.ReadSecret(key, true); err != nil {
				return nil, err
			}
			results[key] = string(res)
		}

	} else {

		secretName := cluster.GetAuthK8sSecret()
		secretNamespace, err := k8sutil.GetThisPodNamespace()
		if err != nil {
			return nil, err
		}
		nn := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
		secret := &corev1.Secret{}
		if err := c.Get(context.TODO(), nn, secret); err != nil {
			return nil, err
		}
		if secret.Data == nil {
			return nil, fmt.Errorf("Provided secret %s is empty", secretName)
		}

		for _, key := range keys {
			res, ok := secret.Data[key]
			if !ok {
				return nil, fmt.Errorf("There is no key %s in secret %s", key, secretName)
			}
			results[key] = string(res)
		}

	}

	return results, nil
}
