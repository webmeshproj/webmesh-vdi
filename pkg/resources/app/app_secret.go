package app

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *AppReconciler) reconcileAppSecret(reqLogger logr.Logger, cluster *v1alpha1.VDICluster) (err error) {
	nn := types.NamespacedName{Name: cluster.GetAppSecretsName(), Namespace: cluster.GetCoreNamespace()}
	found := &corev1.Secret{}
	if err := r.client.Get(context.TODO(), nn, found); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// We are generating a password
		reqLogger.Info("Generating jwt signing key and creating new secret", "Secret.Name", nn.Name, "Secret.Namespace", nn.Namespace)
		passw := common.GeneratePassword(32)
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            nn.Name,
				Namespace:       nn.Namespace,
				Labels:          cluster.GetComponentLabels("app-secret"),
				OwnerReferences: cluster.OwnerReferences(),
			},
			Data: map[string][]byte{
				v1alpha1.JWTSecretKey: []byte(passw),
			},
		}
		if err := r.client.Create(context.TODO(), secret); err != nil {
			return err
		}
		return nil
	}

	if _, ok := found.Data[v1alpha1.JWTSecretKey]; !ok {
		reqLogger.Info("No jwt signing key found, generating new one", "Secret.Name", nn.Name, "Secret.Namespace", nn.Namespace)
		passw := common.GeneratePassword(32)
		found.Data[v1alpha1.JWTSecretKey] = []byte(passw)
		if err := r.client.Update(context.TODO(), found); err != nil {
			return err
		}
	}
	return nil
}
