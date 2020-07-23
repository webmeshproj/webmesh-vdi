package app

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const passwordKey = "password"

func (r *Reconciler) reconcileAdminSecret(reqLogger logr.Logger, cluster *v1alpha1.VDICluster) (password string, err error) {
	nn := types.NamespacedName{Name: cluster.GetAdminSecret(), Namespace: cluster.GetCoreNamespace()}
	found := &corev1.Secret{}
	if err := r.client.Get(context.TODO(), nn, found); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return "", err
		}
		// We are generating a password
		reqLogger.Info("Generating password and creating new admin secret", "Secret.Name", nn.Name, "Secret.Namespace", nn.Namespace)
		passw := common.GeneratePassword(16)
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            nn.Name,
				Namespace:       nn.Namespace,
				Labels:          cluster.GetComponentLabels("admin-secret"),
				OwnerReferences: cluster.OwnerReferences(),
			},
			Data: map[string][]byte{
				passwordKey: []byte(passw),
			},
		}
		if err := r.client.Create(context.TODO(), secret); err != nil {
			return "", err
		}
		return passw, nil
	}
	existingPassw, ok := found.Data[passwordKey]
	if !ok {
		// delete the secret and requeue, currently migration depends on the admin
		// password - but long-term this is probably not a good idea
		if err := r.client.Delete(context.TODO(), found); err != nil {
			return "", err
		}
		return "", errors.NewRequeueError("Regenerating admin secret", 3)
	}
	return string(existingPassw), nil
}
