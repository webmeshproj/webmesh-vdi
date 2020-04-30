package local

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (a *LocalAuthProvider) getSecret() (*corev1.Secret, error) {
	nn := types.NamespacedName{Name: a.cluster.GetAppSecretsName(), Namespace: a.cluster.GetCoreNamespace()}
	secret := &corev1.Secret{}
	return secret, a.client.Get(context.TODO(), nn, secret)
}

func (a *LocalAuthProvider) getPasswdFile() (io.ReadWriter, error) {
	secret, err := a.getSecret()
	if err != nil {
		return nil, err
	}

	data, ok := secret.Data[passwdKey]
	if !ok {
		return nil, fmt.Errorf("No %s in the app secret", passwdKey)
	}
	return bytes.NewBuffer(data), nil
}

func (a *LocalAuthProvider) updatePasswdFile(rdr io.Reader) error {
	body, err := ioutil.ReadAll(rdr)
	if err != nil {
		return err
	}
	secret, err := a.getSecret()
	if err != nil {
		return err
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[passwdKey] = body
	return a.client.Update(context.TODO(), secret)
}
