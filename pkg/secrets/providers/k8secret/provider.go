package k8secret

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provider implements a SecretsProvider that matches secret names to
// keys in a single configured secret.
type Provider struct {
	common.SecretsProvider

	// the k8s client
	client client.Client
	// the name of the secret backing this engine
	secretName types.NamespacedName
}

// Blank assignmnt to make sure Provider satisfies the SecretsProvider
// interface.
var _ common.SecretsProvider = &Provider{}

// New returns a new Provider.
func New() *Provider {
	return &Provider{}
}

// Setup will set the client inteface and secret name, and then ensure the presence
// of the secret in the cluster.
func (k *Provider) Setup(client client.Client, cluster *v1alpha1.VDICluster) error {
	k.secretName = types.NamespacedName{Name: cluster.GetAppSecretsName(), Namespace: cluster.GetCoreNamespace()}
	k.client = client
	return k.ensureSecret(cluster)
}

// ensureSecret makes sure the configured secret exists in the cluster.
func (k *Provider) ensureSecret(cluster *v1alpha1.VDICluster) error {
	if _, err := k.getSecret(); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            k.secretName.Name,
				Namespace:       k.secretName.Namespace,
				Labels:          cluster.GetComponentLabels("app-secret"),
				OwnerReferences: cluster.OwnerReferences(),
			},
		}
		return k.client.Create(context.TODO(), secret)
	}
	return nil
}

// getSecret will retrieve the configured secret.
func (k *Provider) getSecret() (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	return secret, k.client.Get(context.TODO(), k.secretName, secret)
}

// ReadSecret returns the data in the key specified by the given name.
func (k *Provider) ReadSecret(name string) ([]byte, error) {
	secret, err := k.getSecret()
	if err != nil {
		return nil, err
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	data, ok := secret.Data[name]
	if !ok {
		return nil, errors.NewSecretNotFoundError(name)
	}
	return data, nil
}

// WriteSecret will write the given data to the key of the given name and then
// update the secret.
func (k *Provider) WriteSecret(name string, content []byte) error {
	secret, err := k.getSecret()
	if err != nil {
		return err
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	if content == nil {
		delete(secret.Data, name)
	} else {
		secret.Data[name] = content
	}
	if err := k.client.Update(context.TODO(), secret); err != nil {
		return err
	}
	return nil
}

// Close just returns nil because no cleanup is necessary.
func (k *Provider) Close() error { return nil }
