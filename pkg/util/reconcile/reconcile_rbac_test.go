package reconcile

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFakeSvcAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-svc-account",
			Namespace: "fake-namespace",
		},
	}
}

func TestReconcileSvcAccount(t *testing.T) {
	c := getFakeClient(t)
	if err := ReconcileServiceAccount(testLogger, c, newFakeSvcAccount()); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := ReconcileServiceAccount(testLogger, c, newFakeSvcAccount()); err != nil {
		t.Error("Expected no error, got:", err)
	}
}

func newFakeClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-cluster-role",
			Namespace: "fake-namespace",
		},
	}
}

func TestReconcileClusterRole(t *testing.T) {
	c := getFakeClient(t)
	role := newFakeClusterRole()
	if err := ReconcileClusterRole(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := ReconcileClusterRole(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
}

func newFakeClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-cluster-role-binding",
			Namespace: "fake-namespace",
		},
	}
}

func TestReconcileClusterRoleBinding(t *testing.T) {
	c := getFakeClient(t)
	role := newFakeClusterRoleBinding()
	if err := ReconcileClusterRoleBinding(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := ReconcileClusterRoleBinding(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
}
