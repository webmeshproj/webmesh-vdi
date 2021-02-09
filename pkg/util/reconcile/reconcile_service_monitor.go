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

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceMonitor reconciles a ServiceMonitor with the cluster.
func ServiceMonitor(ctx context.Context, reqLogger logr.Logger, c client.Client, sm *promv1.ServiceMonitor) error {
	if err := k8sutil.SetCreationSpecAnnotation(&sm.ObjectMeta, sm); err != nil {
		return err
	}
	found := &promv1.ServiceMonitor{}
	if err := c.Get(ctx, types.NamespacedName{Name: sm.GetName(), Namespace: sm.GetNamespace()}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the ServiceMonitor
		reqLogger.Info("Creating new ServiceMonitor ", "Name", sm.Name, "Namespace", sm.Namespace)
		if err := c.Create(ctx, sm); err != nil {
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
		return c.Update(ctx, found)
	}

	return nil
}
