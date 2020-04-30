package local

import (
	"context"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const passwdKey = "passwd"

func (l *LocalAuthProvider) Reconcile(reqLogger logr.Logger, c client.Client, cluster *v1alpha1.VDICluster, adminPass string) error {
	adminRole := cluster.GetAdminRole()
	hash, err := common.HashPassword(adminPass)
	if err != nil {
		return err
	}

	nn := types.NamespacedName{Name: cluster.GetAppSecretsName(), Namespace: cluster.GetCoreNamespace()}
	secret := &corev1.Secret{}
	if err := c.Get(context.TODO(), nn, secret); err != nil {
		return err
	}
	if _, ok := secret.Data[passwdKey]; !ok {
		secret.Data[passwdKey] = []byte(fmt.Sprintf("admin:%s:%s\n", adminRole.GetName(), hash))
		if err := c.Update(context.TODO(), secret); err != nil {
			return err
		}
	} // else check if password is correct or no?
	return nil
}
