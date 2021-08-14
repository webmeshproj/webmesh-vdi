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

package desktop

import (
	"context"

	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	"github.com/kvdi/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Reconciler) reclaimVolumes(reqLogger logr.Logger, instance *desktopsv1.Session) error {
	cluster, err := instance.GetVDICluster(f.client)
	if err != nil {
		return err
	}
	if cluster.GetUserdataVolumeSpec() != nil {

		pod := &corev1.Pod{}
		if err := f.client.Get(context.TODO(), types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, pod); err == nil {
			reqLogger.Info("Pod still exists, sending delete and requeueing")
			if err := f.client.Delete(context.TODO(), pod); err != nil {
				return err
			}
			return errors.NewRequeueError("Desktop pod is still terminating", 3)
		} else if client.IgnoreNotFound(err) != nil {
			return err
		}

		volMapCM, err := f.getVolMapForCluster(cluster)
		if err != nil {
			return err
		}

		if volMapCM.Data == nil {
			reqLogger.Info("The userdata configmap is empty, skipping finalizer")
			return nil
		}

		pvName, ok := volMapCM.Data[instance.GetUser()]
		if !ok {
			reqLogger.Info("There is no key for this user in the userdata configmap, skipping finalizer")
			return nil
		}

		if pvc, err := f.getPVCForInstance(cluster, instance); err == nil {
			if err = f.client.Delete(context.TODO(), pvc); err != nil {
				reqLogger.Error(err, "Tried to send a delete and failed")
			}
			return errors.NewRequeueError("PVC is still being terminated", 5)
		} else if client.IgnoreNotFound(err) != nil {
			return err
		}

		pv, err := f.getPV(pvName)
		if err != nil {
			if client.IgnoreNotFound(err) == nil {
				reqLogger.Info("The persistent volume for this user has been deleted, skipping finalizer")
				return nil
			}
			return err
		}

		reqLogger.Info("Freeing pv from old pvc claim")
		if changed, err := f.freePV(pv); err != nil {
			return err
		} else if changed {
			return errors.NewRequeueError("Making sure our PV is free", 5)
		}
	}

	return nil
}
