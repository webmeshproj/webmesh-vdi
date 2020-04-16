package reconcile

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/tinyzimmer/kvdi/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileServiceAccount(reqLogger logr.Logger, c client.Client, acct *corev1.ServiceAccount) error {
	found := &corev1.ServiceAccount{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: acct.Name, Namespace: acct.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the service account
		reqLogger.Info("Creating new service account", "ServiceAccount.Name", acct.Name, "ServiceAccount.Namespace", acct.Namespace)
		if err := c.Create(context.TODO(), acct); err != nil {
			return err
		}
	}
	return nil
}

func ReconcileClusterRoleBinding(reqLogger logr.Logger, c client.Client, roleBinding *rbacv1.ClusterRoleBinding) error {
	if err := util.SetCreationSpecAnnotation(&roleBinding.ObjectMeta, roleBinding); err != nil {
		return err
	}
	found := &rbacv1.ClusterRoleBinding{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the role binding
		reqLogger.Info("Creating new cluster role binding", "Name", roleBinding.Name)
		if err := c.Create(context.TODO(), roleBinding); err != nil {
			return err
		}
		return nil
	}

	// Check the found role binding spec
	if !util.CreationSpecsEqual(roleBinding.ObjectMeta, found.ObjectMeta) {
		// We need to update the role binding
		reqLogger.Info("Role binding annotation spec has changed, updating", "Name", roleBinding.Name)
		found.Subjects = roleBinding.Subjects
		found.RoleRef = roleBinding.RoleRef
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
		return nil
	}

	return nil
}
