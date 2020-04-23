package v1alpha1

import (
	"context"
	"fmt"

	"github.com/tinyzimmer/kvdi/version"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetTemplate retrieves the DesktopTemplate for this Desktop instance.
func (d *Desktop) GetTemplate(c client.Client) (*DesktopTemplate, error) {
	nn := types.NamespacedName{Name: d.Spec.Template, Namespace: metav1.NamespaceAll}
	found := &DesktopTemplate{}
	return found, c.Get(context.TODO(), nn, found)
}

// GetVDICluster retrieves the VDICluster for this Desktop instance
func (d *Desktop) GetVDICluster(c client.Client) (*VDICluster, error) {
	nn := types.NamespacedName{Name: d.Spec.VDICluster, Namespace: metav1.NamespaceAll}
	found := &VDICluster{}
	return found, c.Get(context.TODO(), nn, found)
}

// GetUser returns the username that should be used inside the instance.
func (d *Desktop) GetUser() string {
	if d.Spec.User == "" {
		return "anonymous"
	}
	return d.Spec.User
}

// GetNoVNCProxyImage returns the novnc-proxy image for the desktop instance.
func (d *Desktop) GetNoVNCProxyImage() string {
	return fmt.Sprintf("quay.io/tinyzimmer/kvdi:novnc-proxy-%s", version.Version)
}

// OwnerReferences returns an owner reference slice with this Desktop
// instance as the owner.
func (d *Desktop) OwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         d.APIVersion,
			Kind:               d.Kind,
			Name:               d.GetName(),
			UID:                d.GetUID(),
			Controller:         &trueVal,
			BlockOwnerDeletion: &falseVal,
		},
	}
}
