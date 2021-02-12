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
	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newDesktopPodForCR(cluster *appv1.VDICluster, tmpl *desktopsv1.Template, instance *desktopsv1.Session, envSecret string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetName(),
			Namespace:       instance.GetNamespace(),
			Labels:          cluster.GetDesktopLabels(instance),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: corev1.PodSpec{
			Hostname:           instance.GetName(),
			Subdomain:          instance.GetName(),
			ServiceAccountName: instance.GetServiceAccount(),
			SecurityContext:    tmpl.GetDesktopPodSecurityContext(),
			Volumes:            k8sutil.GetDesktopVolumesFromTemplate(tmpl, cluster, instance),
			ImagePullSecrets:   tmpl.GetDesktopPullSecrets(),
			InitContainers:     tmpl.GetInitContainers(),
			Containers: []corev1.Container{
				tmpl.GetDesktopProxyContainer(),
				{
					Name:            "desktop",
					Image:           tmpl.GetDesktopImage(),
					ImagePullPolicy: tmpl.GetDesktopPullPolicy(),
					VolumeMounts:    k8sutil.GetDesktopVolumeMountsFromTemplate(tmpl, cluster, instance),
					VolumeDevices:   tmpl.GetVolumeDevices(),
					SecurityContext: tmpl.GetDesktopContainerSecurityContext(),
					Env:             tmpl.GetDesktopEnvVars(instance),
					Lifecycle:       tmpl.GetLifecycle(),
					Resources:       tmpl.GetDesktopResources(),
				},
			},
		},
	}
	if envSecret != "" {
		pod.Spec.Containers[1].EnvFrom = []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: envSecret,
					},
				},
			},
		}
	}
	if tmpl.DindIsEnabled() {
		pod.Spec.Containers = append(pod.Spec.Containers, tmpl.GetDindContainer())
	}
	return pod
}

func newServiceForCR(cluster *appv1.VDICluster, instance *desktopsv1.Session) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetName(),
			Namespace:       instance.GetNamespace(),
			Labels:          cluster.GetDesktopLabels(instance),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: cluster.GetDesktopLabels(instance),
			Ports: []corev1.ServicePort{
				{
					Name:       "kvdi-proxy",
					Port:       v1.WebPort,
					TargetPort: intstr.FromInt(v1.WebPort),
				},
			},
		},
	}
}
