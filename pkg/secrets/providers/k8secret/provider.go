package k8secret

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8SecretProvider struct {
	v1alpha1.SecretsProvider

	client     client.Client
	secretName types.NamespacedName
}

var _ v1alpha1.SecretsProvider = &K8SecretProvider{}

func New() *K8SecretProvider {
	return &K8SecretProvider{}
}

func (k *K8SecretProvider) Setup(client client.Client, cluster *v1alpha1.VDICluster) error {
	k.secretName = types.NamespacedName{Name: cluster.GetAppSecretsName(), Namespace: cluster.GetCoreNamespace()}
	k.client = client
	return k.ensureSecret(cluster)
}

func (k *K8SecretProvider) ensureSecret(cluster *v1alpha1.VDICluster) error {
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

func (k *K8SecretProvider) getSecret() (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	return secret, k.client.Get(context.TODO(), k.secretName, secret)
}

func (k *K8SecretProvider) GetName() string {
	return k.secretName.Name
}

func (k *K8SecretProvider) ReadSecret(name string) ([]byte, error) {
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

func (k *K8SecretProvider) WriteSecret(name string, content []byte) error {
	secret, err := k.getSecret()
	if err != nil {
		return err
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[name] = content
	if err := k.client.Update(context.TODO(), secret); err != nil {
		return err
	}
	return nil
}
