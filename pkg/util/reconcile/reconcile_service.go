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

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service will reconcile a provided service spec with the cluster.
func Service(ctx context.Context, reqLogger logr.Logger, c client.Client, svc *corev1.Service) error {
	if err := k8sutil.SetCreationSpecAnnotation(&svc.ObjectMeta, svc); err != nil {
		return err
	}
	found := &corev1.Service{}
	if err := c.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the service
		reqLogger.Info("Creating new service", "Service.Name", svc.Name, "Service.Namespace", svc.Namespace)
		if err := c.Create(ctx, svc); err != nil {
			return err
		}
		return nil
	}

	// Check the found service spec
	if !k8sutil.CreationSpecsEqual(svc.ObjectMeta, found.ObjectMeta) {
		// We need to update the service
		reqLogger.Info("Service annotation spec has changed, deleting and requeing", "Service.Name", svc.Name, "Service.Namespace", svc.Namespace)
		if err := c.Delete(ctx, found); err != nil {
			return err
		}
		return errors.NewRequeueError("Deleted service definition, requeueing", 2)
	}

	return nil
}
