/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package k8secret

import (
	"context"
	"encoding/base64"
	"encoding/json"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
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
func (k *Provider) Setup(client client.Client, cluster *appv1.VDICluster) error {
	k.secretName = types.NamespacedName{Name: cluster.GetAppSecretsName(), Namespace: cluster.GetCoreNamespace()}
	k.client = client
	return k.ensureSecret(cluster)
}

// ensureSecret makes sure the configured secret exists in the cluster.
func (k *Provider) ensureSecret(cluster *appv1.VDICluster) error {
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

// ReadSecretMap implements SecretsProvider and returns a stored map secret.
func (k *Provider) ReadSecretMap(name string) (map[string][]byte, error) {
	contents, err := k.ReadSecret(name)
	if err != nil {
		return nil, err
	}
	// json marshalled bytes are encoded with base64
	outEncoded := make(map[string]string)
	if err := json.Unmarshal(contents, &outEncoded); err != nil {
		return nil, err
	}
	out := make(map[string][]byte)
	for k, v := range outEncoded {
		vBytes, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		out[k] = vBytes
	}
	return out, nil
}

// WriteSecretMap implements SecretsProvider and will write the key-value pair
// to the secrets backend. The secret can be read back in the same fashion.
// This will be the preferred function going forward.
func (k *Provider) WriteSecretMap(name string, content map[string][]byte) error {
	if content == nil {
		return k.WriteSecret(name, nil)
	}
	// json will base64 encode the byte slices
	out, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return k.WriteSecret(name, out)
}

// Close just returns nil because no cleanup is necessary.
func (k *Provider) Close() error { return nil }
