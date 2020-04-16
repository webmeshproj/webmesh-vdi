package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *VDICluster) GetCoreNamespace() string {
	if c.Spec.AppNamespace != "" {
		return c.Spec.AppNamespace
	}
	return defaultNamespace
}

func (c *VDICluster) GetPullSecrets() []corev1.LocalObjectReference {
	return c.Spec.ImagePullSecrets
}

func (c *VDICluster) GetComponentLabels(component string) map[string]string {
	labels := c.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[VDIClusterLabel] = c.GetName()
	labels[ComponentLabel] = component
	return labels
}

func (c *VDICluster) GetDesktopLabels(desktop *Desktop) map[string]string {
	labels := desktop.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[UserLabel] = desktop.Spec.User
	labels[VDIClusterLabel] = c.GetName()
	labels[ComponentLabel] = "desktop"
	labels[DesktopNameLabel] = desktop.GetName()
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
			Controller:         &trueVal,
			BlockOwnerDeletion: &trueVal,
		},
	}
}
