package reconcile

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccount will ensure a service account in the cluster
func ServiceAccount(reqLogger logr.Logger, c client.Client, acct *corev1.ServiceAccount) error {
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

// ClusterRole will ensure a ClusterRole with the cluster.
func ClusterRole(reqLogger logr.Logger, c client.Client, role *rbacv1.ClusterRole) error {
	if err := k8sutil.SetCreationSpecAnnotation(&role.ObjectMeta, role); err != nil {
		return err
	}
	found := &rbacv1.ClusterRole{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the role
		reqLogger.Info("Creating new cluster role ", "Name", role.Name)
		if err := c.Create(context.TODO(), role); err != nil {
			return err
		}
		return nil
	}

	// Check the found role spec
	if !k8sutil.CreationSpecsEqual(role.ObjectMeta, found.ObjectMeta) {
		// We need to update the role
		reqLogger.Info("Role annotation spec has changed, updating", "Name", role.Name)
		found.Rules = role.Rules
		found.SetAnnotations(role.GetAnnotations())
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
		return nil
	}

	return nil
}

// ClusterRoleBinding will ensure a cluster role binding.
func ClusterRoleBinding(reqLogger logr.Logger, c client.Client, roleBinding *rbacv1.ClusterRoleBinding) error {
	if err := k8sutil.SetCreationSpecAnnotation(&roleBinding.ObjectMeta, roleBinding); err != nil {
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
	if !k8sutil.CreationSpecsEqual(roleBinding.ObjectMeta, found.ObjectMeta) {
		// We need to update the role binding
		reqLogger.Info("Role binding annotation spec has changed, updating", "Name", roleBinding.Name)
		found.Subjects = roleBinding.Subjects
		found.RoleRef = roleBinding.RoleRef
		found.SetAnnotations(roleBinding.GetAnnotations())
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
		return nil
	}

	return nil
}

// VDIRole reconciles a VDIRole with the cluster.
func VDIRole(reqLogger logr.Logger, c client.Client, role *v1alpha1.VDIRole) error {
	if err := k8sutil.SetCreationSpecAnnotation(&role.ObjectMeta, role); err != nil {
		return err
	}
	found := &v1alpha1.VDIRole{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: metav1.NamespaceAll}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the role
		reqLogger.Info("Creating new VDI role ", "Name", role.Name)
		if err := c.Create(context.TODO(), role); err != nil {
			return err
		}
		return nil
	}

	// Check the found role spec
	if !k8sutil.CreationSpecsEqual(role.ObjectMeta, found.ObjectMeta) {
		// We need to update the role
		reqLogger.Info("Role annotation spec has changed, updating", "Name", role.Name)
		found.Rules = role.Rules
		found.SetLabels(role.GetLabels())
		found.SetAnnotations(role.GetAnnotations())
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
		return nil
	}

	return nil
}
