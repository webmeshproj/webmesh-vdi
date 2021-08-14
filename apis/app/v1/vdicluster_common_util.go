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
	"fmt"
	"reflect"

	v1 "github.com/kvdi/kvdi/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetCoreNamespace returns the namespace where kVDI components should be created.
func (c *VDICluster) GetCoreNamespace() string {
	if c.Spec.AppNamespace != "" {
		return c.Spec.AppNamespace
	}
	return v1.DefaultNamespace
}

// GetAppServiceType returns the type of service to create in front of the app pods.
func (c *VDICluster) GetAppServiceType() corev1.ServiceType {
	if c.Spec.App != nil && c.Spec.App.ServiceType != "" {
		return c.Spec.App.ServiceType
	}
	return corev1.ServiceTypeLoadBalancer
}

// NamespacedName returns the NamespacedName of this VDICluster.
func (c *VDICluster) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: c.GetName(), Namespace: metav1.NamespaceAll}
}

// GetPullSecrets returns any pull secrets required for pulling images.
func (c *VDICluster) GetPullSecrets() []corev1.LocalObjectReference {
	return c.Spec.ImagePullSecrets
}

// GetComponentLabels returns the labels to apply to a given kVDI component.
func (c *VDICluster) GetComponentLabels(component string) map[string]string {
	labels := c.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[v1.VDIClusterLabel] = c.GetName()
	labels[v1.ComponentLabel] = component
	return labels
}

// GetClusterDesktopsSelector gets the label selector for looking up all desktops
// owned by this VDICluster.
func (c *VDICluster) GetClusterDesktopsSelector() client.MatchingLabels {
	return client.MatchingLabels{
		v1.VDIClusterLabel: c.GetName(),
	}
}

// GetUserDesktopsSelector gets the label selector to use for looking up a user's
// desktop sessions.
func (c *VDICluster) GetUserDesktopsSelector(username string) client.MatchingLabels {
	return client.MatchingLabels{
		v1.UserLabel:       username,
		v1.VDIClusterLabel: c.GetName(),
	}
}

// OwnerReferences returns an owner reference slice with this VDICluster
// instance as the owner.
func (c *VDICluster) OwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         c.APIVersion,
			Kind:               c.Kind,
			Name:               c.GetName(),
			UID:                c.GetUID(),
			Controller:         &v1.True,
			BlockOwnerDeletion: &v1.False,
		},
	}
}

// GetUserdataSelector returns the selector to use for locating PVCs for a user's $HOME.
func (c *VDICluster) GetUserdataSelector() *UserdataSelector {
	return c.Spec.UserdataSelector
}

// GetUserdataVolumeSpec returns the spec for creating PVCs for user persistence.
func (c *VDICluster) GetUserdataVolumeSpec() *corev1.PersistentVolumeClaimSpec {
	if c.Spec.UserdataSpec != nil && !reflect.DeepEqual(*c.Spec.UserdataSpec, corev1.PersistentVolumeClaimSpec{}) {
		return c.Spec.UserdataSpec
	}
	return nil
}

// GetUserdataVolumeName returns the name of the userdata volume for the given user.
func (c *VDICluster) GetUserdataVolumeName(username string) string {
	return fmt.Sprintf("%s-%s-userdata", c.GetName(), username)
}

// GetUserdataVolumeMapName returns the name of the configmap where user's are mapped to PVs.
func (c *VDICluster) GetUserdataVolumeMapName() types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-userdata-volume-map", c.GetName()),
		Namespace: c.GetCoreNamespace(),
	}
}
