package reconcile

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileService will reconcile a provided service spec with the cluster.
func ReconcileService(reqLogger logr.Logger, c client.Client, svc *corev1.Service) error {
	if err := util.SetCreationSpecAnnotation(&svc.ObjectMeta, svc); err != nil {
		return err
	}
	found := &corev1.Service{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the service
		reqLogger.Info("Creating new service", "Service.Name", svc.Name, "Service.Namespace", svc.Namespace)
		if err := c.Create(context.TODO(), svc); err != nil {
			return err
		}
		return nil
	}

	// Check the found service spec
	if !util.CreationSpecsEqual(svc.ObjectMeta, found.ObjectMeta) {
		// We need to update the service
		reqLogger.Info("Service annotation spec has changed, deleting and requeing", "Service.Name", svc.Name, "Service.Namespace", svc.Namespace)
		if err := c.Delete(context.TODO(), found); err != nil {
			return err
		}
		return errors.NewRequeueError("Deleted service definition, requeueing", 2)
	}

	return nil
}
