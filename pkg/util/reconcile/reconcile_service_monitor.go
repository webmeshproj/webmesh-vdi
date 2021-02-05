package reconcile

import (
	"context"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceMonitor reconciles a ServiceMonitor with the cluster.
func ServiceMonitor(reqLogger logr.Logger, c client.Client, sm *promv1.ServiceMonitor) error {
	if err := k8sutil.SetCreationSpecAnnotation(&sm.ObjectMeta, sm); err != nil {
		return err
	}
	found := &promv1.ServiceMonitor{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: sm.GetName(), Namespace: sm.GetNamespace()}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the ServiceMonitor
		reqLogger.Info("Creating new ServiceMonitor ", "Name", sm.Name, "Namespace", sm.Namespace)
		if err := c.Create(context.TODO(), sm); err != nil {
			return err
		}
		return nil
	}

	// Check the found ServiceMonitor spec
	if !k8sutil.CreationSpecsEqual(sm.ObjectMeta, found.ObjectMeta) {
		// We need to update the role
		reqLogger.Info("ServiceMonitor annotation spec has changed, updating", "Name", sm.Name, "Namespace", sm.Namespace)
		found.Spec = sm.Spec
		found.SetLabels(sm.GetLabels())
		found.SetAnnotations(sm.GetAnnotations())
		return c.Update(context.TODO(), found)
	}

	return nil
}
