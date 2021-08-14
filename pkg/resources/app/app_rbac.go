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
	appv1 "github.com/kvdi/kvdi/apis/app/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var verbsReadOnly = []string{"get", "list", "watch"}
var verbsAll = []string{"create", "delete", "get", "list", "patch", "update", "watch"}

var appRules = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"app.kvdi.io"},
		Resources: []string{"vdiclusters"},
		Verbs:     verbsReadOnly,
	},
	{
		APIGroups: []string{"rbac.kvdi.io"},
		Resources: []string{"vdiroles"},
		Verbs:     verbsAll,
	},
	{
		APIGroups: []string{"desktops.kvdi.io"},
		Resources: []string{"sessions", "templates"},
		Verbs:     verbsAll,
	},
	{
		APIGroups: []string{""},
		Resources: []string{"pods", "pods/log", "services", "namespaces", "endpoints", "serviceaccounts"},
		Verbs:     verbsReadOnly,
	},
	{
		APIGroups: []string{""},
		Resources: []string{"configmaps", "secrets"},
		Verbs:     verbsAll,
	},
}

func newAppClusterRoleForCR(instance *appv1.VDICluster) *rbacv1.ClusterRole {
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

func newAppServiceAccountForCR(instance *appv1.VDICluster) *corev1.ServiceAccount {
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

func newRoleBindingsForCR(instance *appv1.VDICluster) *rbacv1.ClusterRoleBinding {
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
