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

package reconcile

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
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
	if err := ServiceAccount(testLogger, c, newFakeSvcAccount()); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := ServiceAccount(testLogger, c, newFakeSvcAccount()); err != nil {
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
	if err := ClusterRole(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := ClusterRole(testLogger, c, role); err != nil {
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
	if err := ClusterRoleBinding(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := ClusterRoleBinding(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
}

func newFakeVDIRole() *v1alpha1.VDIRole {
	return &v1alpha1.VDIRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-cluster-role-binding",
		},
	}
}

func TestReconcileVDIRole(t *testing.T) {
	c := getFakeClient(t)
	role := newFakeVDIRole()
	if err := VDIRole(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := VDIRole(testLogger, c, role); err != nil {
		t.Error("Expected no error, got:", err)
	}
}
