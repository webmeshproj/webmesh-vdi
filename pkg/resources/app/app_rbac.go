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

package app

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var verbsReadOnly = []string{"get", "list", "watch"}

var appRules = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"kvdi.io"},
		Resources: []string{rbacv1.ResourceAll},
		Verbs:     []string{rbacv1.VerbAll},
	},
	{
		APIGroups: []string{""},
		Resources: []string{"pods", "pods/log", "services", "namespaces", "endpoints", "serviceaccounts"},
		Verbs:     verbsReadOnly,
	},
	{
		APIGroups: []string{""},
		Resources: []string{"configmaps", "secrets"},
		Verbs:     []string{rbacv1.VerbAll},
	},
}

func newAppClusterRoleForCR(instance *v1alpha1.VDICluster) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       metav1.NamespaceAll,
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Rules: appRules,
	}
}

func newAppServiceAccountForCR(instance *v1alpha1.VDICluster) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
	}
}

func newRoleBindingsForCR(instance *v1alpha1.VDICluster) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       metav1.NamespaceAll,
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      instance.GetAppName(),
				Namespace: instance.GetCoreNamespace(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     instance.GetAppName(),
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}
