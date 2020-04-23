package reconcile

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/util"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileConfigMap will reconcile a provided configmap with the cluster.
func ReconcileConfigMap(reqLogger logr.Logger, c client.Client, cm *corev1.ConfigMap) error {
	if err := util.SetCreationSpecAnnotation(&cm.ObjectMeta, cm); err != nil {
		return err
	}
	found := &corev1.ConfigMap{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the config map
		reqLogger.Info("Creating new ConfigMap", "ConfigMap.Name", cm.Name, "ConfigMap.Namespace", cm.Namespace)
		if err := c.Create(context.TODO(), cm); err != nil {
			return err
		}
		return nil
	}

	// Check the found service spec
	if !util.CreationSpecsEqual(cm.ObjectMeta, found.ObjectMeta) {
		// We need to update the configmap
		reqLogger.Info("ConfigMap annotation spec has changed, updating", "ConfigMap.Name", cm.Name, "ConfigMap.Namespace", cm.Namespace)
		found.Data = cm.Data
		found.SetAnnotations(cm.GetAnnotations())
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
	}

	return nil
}
