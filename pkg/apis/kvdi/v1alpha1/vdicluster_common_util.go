package v1alpha1

import (
	"fmt"
	"reflect"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

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

// GetUserDesktopLabels returns the labels to apply to the pod for a user's desktop session.
func (c *VDICluster) GetUserDesktopLabels(username string) map[string]string {
	return map[string]string{
		v1.UserLabel:       username,
		v1.VDIClusterLabel: c.GetName(),
	}
}

// GetDesktopLabels returns desktop labels.
// TODO: Find out if this or GetUserDesktopLabels is actually being used.
func (c *VDICluster) GetDesktopLabels(desktop *Desktop) map[string]string {
	labels := desktop.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[v1.UserLabel] = desktop.Spec.User
	labels[v1.VDIClusterLabel] = c.GetName()
	labels[v1.ComponentLabel] = "desktop"
	labels[v1.DesktopNameLabel] = desktop.GetName()
	return labels
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
			Controller:         &v1.TrueVal,
			BlockOwnerDeletion: &v1.FalseVal,
		},
	}
}

// GetUserdataVolumeSpec returns the spec for creating PVCs for user persistence.
func (c *VDICluster) GetUserdataVolumeSpec() *corev1.PersistentVolumeClaimSpec {
	if c.Spec.UserDataSpec != nil && !reflect.DeepEqual(*c.Spec.UserDataSpec, corev1.PersistentVolumeClaimSpec{}) {
		return c.Spec.UserDataSpec
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
