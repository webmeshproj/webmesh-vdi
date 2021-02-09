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

// PersistentVolumeClaim will reconcile a persistent volume with the kubernetes
// cluster. If it exists, we do nothing for now.
func PersistentVolumeClaim(ctx context.Context, reqLogger logr.Logger, c client.Client, pvc *corev1.PersistentVolumeClaim) error {
	// Set the creation spec anyway so it's there if we need it in the future
	if err := k8sutil.SetCreationSpecAnnotation(&pvc.ObjectMeta, pvc); err != nil {
		return err
	}

	found := &corev1.PersistentVolumeClaim{}
	if err := c.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the volume
		reqLogger.Info("Creating new Persistent Volume Claim", "PV.Name", pvc.Name, "PV.Namespace", pvc.Namespace)
		if err := c.Create(ctx, pvc); err != nil {
			return err
		}
	}

	return nil
}
