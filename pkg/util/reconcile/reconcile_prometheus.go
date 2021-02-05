package reconcile

import (
	"context"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Prometheus reconciles a Prometheus CR with the cluster.
func Prometheus(reqLogger logr.Logger, c client.Client, prom *promv1.Prometheus) error {
	if err := k8sutil.SetCreationSpecAnnotation(&prom.ObjectMeta, prom); err != nil {
		return err
	}
	found := &promv1.Prometheus{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: prom.GetName(), Namespace: prom.GetNamespace()}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the Prometheus CR
		reqLogger.Info("Creating new Prometheus CR", "Name", prom.Name, "Namespace", prom.Namespace)
		if err := c.Create(context.TODO(), prom); err != nil {
			return err
		}
		return nil
	}

	// Check the found Prometheus spec
	if !k8sutil.CreationSpecsEqual(prom.ObjectMeta, found.ObjectMeta) {
		// We need to update the role
		reqLogger.Info("Prometheus annotation spec has changed, updating", "Name", prom.Name, "Namespace", prom.Namespace)
		found.Spec = prom.Spec
		found.SetLabels(prom.GetLabels())
		found.SetAnnotations(prom.GetAnnotations())
		return c.Update(context.TODO(), found)
	}

	return nil
}
