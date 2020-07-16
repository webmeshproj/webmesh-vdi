package ldap

import (
	"context"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getCredentials returns the bind credentials for the configured service account.
func (a *AuthProvider) getCredentials() (user, passw string, err error) {

	userKey := a.cluster.GetLDAPUserDNKey()
	passKey := a.cluster.GetLDAPPasswordKey()

	var res []byte

	if a.cluster.AuthIsUsingSecretEngine() {

		if res, err = a.secrets.ReadSecret(userKey, true); err != nil {
			return
		}
		user = string(res)
		if res, err = a.secrets.ReadSecret(passKey, true); err != nil {
			return
		}
		passw = string(res)

	} else {

		var secretName, secretNamespace string
		secretName = a.cluster.GetAuthK8sSecret()
		secretNamespace, err = k8sutil.GetThisPodNamespace()
		if err != nil {
			return
		}
		nn := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
		secret := &corev1.Secret{}
		if err = a.client.Get(context.TODO(), nn, secret); err != nil {
			return
		}
		if secret.Data == nil {
			err = fmt.Errorf("Provided secret %s is empty", secretName)
			return
		}

		var ok bool

		if res, ok = secret.Data[userKey]; !ok {
			err = fmt.Errorf("There is no key %s in secret %s", userKey, secretName)
			return
		}
		user = string(res)
		if res, ok = secret.Data[passKey]; !ok {
			err = fmt.Errorf("There is no key %s in secret %s", passKey, secretName)
			return
		}
		passw = string(res)

	}

	return
}
