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

package reconcile

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMap reconciles a provided configmap with the cluster.
func ConfigMap(reqLogger logr.Logger, c client.Client, cm *corev1.ConfigMap) error {
	if err := k8sutil.SetCreationSpecAnnotation(&cm.ObjectMeta, cm); err != nil {
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
	if !k8sutil.CreationSpecsEqual(cm.ObjectMeta, found.ObjectMeta) {
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
