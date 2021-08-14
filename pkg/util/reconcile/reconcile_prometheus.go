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

	"github.com/kvdi/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Prometheus reconciles a Prometheus CR with the cluster.
func Prometheus(ctx context.Context, reqLogger logr.Logger, c client.Client, prom *promv1.Prometheus) error {
	if err := k8sutil.SetCreationSpecAnnotation(&prom.ObjectMeta, prom); err != nil {
		return err
	}
	found := &promv1.Prometheus{}
	if err := c.Get(ctx, types.NamespacedName{Name: prom.GetName(), Namespace: prom.GetNamespace()}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the Prometheus CR
		reqLogger.Info("Creating new Prometheus CR", "Name", prom.Name, "Namespace", prom.Namespace)
		if err := c.Create(ctx, prom); err != nil {
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
		return c.Update(ctx, found)
	}

	return nil
}
