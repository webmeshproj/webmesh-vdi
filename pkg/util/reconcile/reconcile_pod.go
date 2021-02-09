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

// Pod will reconcile a given pod definition with the cluster. If a pod
// with the same name exists but has a different configuration, the pod will be
// deleted and requeued. If the found pod has a deletion timestamp (e.g. it is still terminating)
// then the request will also be requued.
func Pod(reqLogger logr.Logger, c client.Client, pod *corev1.Pod) (bool, error) {
	if err := k8sutil.SetCreationSpecAnnotation(&pod.ObjectMeta, pod); err != nil {
		return false, err
	}
	found := &corev1.Pod{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return false, err
		}
		// Create the Pod
		reqLogger.Info("Creating new Pod", "Pod.Name", pod.Name, "Pod.Namespace", pod.Namespace)
		if err := c.Create(context.TODO(), pod); err != nil {
			return false, err
		}
		return true, nil
	}

	// Check if the found pod is in the middle of terminating
	if found.GetDeletionTimestamp() != nil {
		return false, errors.NewRequeueError("Existing pod is still being terminated, requeuing", 3)
	}

	// Check the found pod spec
	if !k8sutil.CreationSpecsEqual(pod.ObjectMeta, found.ObjectMeta) {
		// We need to delete the pod and return a requeue
		if err := c.Delete(context.TODO(), found); err != nil {
			return false, err
		}
		return false, errors.NewRequeueError("Pod spec has changed, recreating", 3)
	}

	return false, nil
}
