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

package v1

import (
	"context"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetTemplate retrieves the DesktopTemplate for this Desktop instance.
func (d *Session) GetTemplate(c client.Client) (*Template, error) {
	nn := types.NamespacedName{Name: d.Spec.Template, Namespace: metav1.NamespaceAll}
	found := &Template{}
	return found, c.Get(context.TODO(), nn, found)
}

// GetServiceAccount returns the service account for this instance.
func (d *Session) GetServiceAccount() string { return d.Spec.ServiceAccount }

// GetUser returns the username that should be used inside the instance.
func (d *Session) GetUser() string {
	if d.Spec.User == "" {
		return "anonymous"
	}
	return d.Spec.User
}

// OwnerReferences returns an owner reference slice with this Desktop
// instance as the owner.
func (d *Session) OwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         d.APIVersion,
			Kind:               d.Kind,
			Name:               d.GetName(),
			UID:                d.GetUID(),
			Controller:         &v1.True,
			BlockOwnerDeletion: &v1.False,
		},
	}
}

// GetVDICluster retrieves the VDICluster for this Desktop instance
func (d *Session) GetVDICluster(c client.Client) (*appv1.VDICluster, error) {
	nn := types.NamespacedName{Name: d.Spec.VDICluster, Namespace: metav1.NamespaceAll}
	found := &appv1.VDICluster{}
	return found, c.Get(context.TODO(), nn, found)
}
