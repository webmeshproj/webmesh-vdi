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

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	"github.com/kvdi/kvdi/pkg/util/errors"
	"github.com/kvdi/kvdi/pkg/util/k8sutil"
	"github.com/kvdi/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Reconciler) reconcileVolumes(ctx context.Context, reqLogger logr.Logger, cluster *appv1.VDICluster, instance *desktopsv1.Session) error {
	volMapCM, err := f.getVolMapForCluster(cluster)
	if err != nil {
		return err
	}
	var existingVol string
	var ok bool
	if existingVol, ok = volMapCM.Data[instance.GetUser()]; ok {
		reqLogger.Info("Fetching existing volume for user")
		if err := f.client.Get(ctx, types.NamespacedName{Name: existingVol, Namespace: metav1.NamespaceAll}, &corev1.PersistentVolume{}); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			reqLogger.Info("The volume referenced in the userdata configmap no longer exists, creating a new one")
			existingVol = ""
		}
	}
	pvc := newPVCForUser(cluster, instance, existingVol)
	return reconcile.PersistentVolumeClaim(ctx, reqLogger, f.client, pvc)
}

func (f *Reconciler) reconcileUserdataMapping(ctx context.Context, reqLogger logr.Logger, cluster *appv1.VDICluster, instance *desktopsv1.Session) error {

	pvc, err := f.getPVCForInstance(cluster, instance)
	if err != nil {
		return err
	}

	if pvc.Spec.VolumeName == "" {
		return errors.NewRequeueError("PVC has not had its volume provisioned yet", 3)
	}

	pvName := pvc.Spec.VolumeName

	pv, err := f.getPV(pvName)
	if err != nil {
		return err
	}

	// it won't harm the running instance and the storage class provider may
	// leave us alone
	if _, err := f.freePV(pv); err != nil {
		return err
	}

	volMapCM, err := f.getVolMapForCluster(cluster)
	if err != nil {
		return err
	}

	if volMapCM.Data == nil {
		volMapCM.Data = make(map[string]string)
	}

	if pv, ok := volMapCM.Data[instance.GetUser()]; !ok || pv != pvName {
		volMapCM.Data[instance.GetUser()] = pvName
		if err := f.client.Update(ctx, volMapCM); err != nil {
			return err
		}
	}

	return nil
}

func newConfigMapForCluster(cluster *appv1.VDICluster) *corev1.ConfigMap {
	nn := cluster.GetUserdataVolumeMapName()
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nn.Name,
			Namespace:       nn.Namespace,
			Labels:          cluster.GetComponentLabels("userdata-map"),
			OwnerReferences: cluster.OwnerReferences(),
		},
		Data: make(map[string]string),
	}
}

func newPVCForUser(cluster *appv1.VDICluster, instance *desktopsv1.Session, existingPVName string) *corev1.PersistentVolumeClaim {
	spec := cluster.GetUserdataVolumeSpec()
	if existingPVName != "" {
		spec.VolumeName = existingPVName
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cluster.GetUserdataVolumeName(instance.GetUser()),
			Namespace:       instance.GetNamespace(),
			Labels:          k8sutil.GetDesktopLabels(cluster, instance),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: *spec,
	}
}
