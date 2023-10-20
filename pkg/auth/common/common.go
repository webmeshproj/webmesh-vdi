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

// Package common contains the core AuthProvider interface and utility functions
// to be used by the auth providers.
package common

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"github.com/kvdi/kvdi/pkg/secrets"
	"github.com/kvdi/kvdi/pkg/util/k8sutil"
)

// GetAuthSecrets is a helper function for retrieving multiple secrets required for
// authentication.
func GetAuthSecrets(c client.Client, cluster *appv1.VDICluster, secrets *secrets.SecretEngine, keys ...string) (map[string]string, error) {
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
			return nil, fmt.Errorf("provided secret %s is empty", secretName)
		}

		for _, key := range keys {
			res, ok := secret.Data[key]
			if !ok {
				return nil, fmt.Errorf("there is no key %s in secret %s", key, secretName)
			}
			results[key] = string(res)
		}
	}
	return results, nil
}
