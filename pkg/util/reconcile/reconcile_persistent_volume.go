package reconcile

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PersistentVolumeClaim will reconcile a persistent volume with the kubernetes
// cluster. If it exists, we do nothing for now.
func PersistentVolumeClaim(reqLogger logr.Logger, c client.Client, pvc *corev1.PersistentVolumeClaim) error {
	// Set the creation spec anyway so it's there if we need it in the future
	if err := k8sutil.SetCreationSpecAnnotation(&pvc.ObjectMeta, pvc.Spec); err != nil {
		return err
	}

	found := &corev1.PersistentVolumeClaim{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the volume
		reqLogger.Info("Creating new Persistent Volume Claim", "PV.Name", pvc.Name, "PV.Namespace", pvc.Namespace)
		if err := c.Create(context.TODO(), pvc); err != nil {
			return err
		}
	}

	return nil
}
