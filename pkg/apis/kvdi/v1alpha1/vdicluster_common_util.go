package v1alpha1

import (
	"fmt"
	"reflect"

	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *VDICluster) GetCoreNamespace() string {
	if c.Spec.AppNamespace != "" {
		return c.Spec.AppNamespace
	}
	return v1.DefaultNamespace
}

func (c *VDICluster) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: c.GetName(), Namespace: metav1.NamespaceAll}
}

func (c *VDICluster) GetPullSecrets() []corev1.LocalObjectReference {
	return c.Spec.ImagePullSecrets
}

func (c *VDICluster) GetComponentLabels(component string) map[string]string {
	labels := c.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[v1.VDIClusterLabel] = c.GetName()
	labels[v1.ComponentLabel] = component
	return labels
}

func (c *VDICluster) GetUserDesktopsSelector(username string) client.MatchingLabels {
	return client.MatchingLabels{
		v1.UserLabel:       username,
		v1.VDIClusterLabel: c.GetName(),
	}
}

func (c *VDICluster) GetUserDesktopLabels(username string) map[string]string {
	return map[string]string{
		v1.UserLabel:       username,
		v1.VDIClusterLabel: c.GetName(),
	}
}

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

func (c *VDICluster) GetUserdataVolumeSpec() *corev1.PersistentVolumeClaimSpec {
	if c.Spec.UserDataSpec != nil && !reflect.DeepEqual(*c.Spec.UserDataSpec, corev1.PersistentVolumeClaimSpec{}) {
		return c.Spec.UserDataSpec
	}
	return nil
}

func (c *VDICluster) GetUserdataVolumeName(username string) string {
	return fmt.Sprintf("%s-%s-userdata", c.GetName(), username)
}

func (c *VDICluster) GetUserdataVolumeMapName() types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-userdata-volume-map", c.GetName()),
		Namespace: c.GetCoreNamespace(),
	}
}
