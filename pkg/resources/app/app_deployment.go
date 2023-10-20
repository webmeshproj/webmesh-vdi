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

package app

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/util/common"
)

func newAppDeploymentForCR(instance *appv1.VDICluster) *appsv1.Deployment {
	containers := []corev1.Container{newAppContainerForCR(instance)}
	volumes := newAppVolumesForCR(instance)
	if instance.RunAppGrafanaSidecar() {
		containers = append(containers, newGrafanaContainerForCR(instance))
		volumes = append(volumes, corev1.Volume{
			Name: "grafana-configs",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-grafana", instance.GetAppName()),
					},
				},
			},
		})
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.GetAppReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: instance.GetComponentLabels("app"),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: instance.GetComponentLabels("app"),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.GetAppName(),
					SecurityContext:    instance.GetAppSecurityContext(),
					Volumes:            volumes,
					ImagePullSecrets:   instance.GetPullSecrets(),
					Containers:         containers,
				},
			},
		},
	}
}

func newAppVolumesForCR(instance *appv1.VDICluster) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "tls-server",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.GetAppServerTLSSecretName(),
				},
			},
		},
		{
			Name: "tls-client",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.GetAppClientTLSSecretName(),
				},
			},
		},
	}
}

func newAppContainerForCR(instance *appv1.VDICluster) corev1.Container {
	args := []string{"--vdi-cluster", instance.GetName()}
	if instance.EnableCORS() {
		args = append(args, "--enable-cors")
	}
	if instance.Spec.App != nil && instance.Spec.App.TLS != nil && instance.Spec.App.TLS.Disable {
		args = append(args, "--disable-tls")
	}
	return corev1.Container{
		Name:            "app",
		Image:           instance.GetAppImage(),
		ImagePullPolicy: instance.GetAppPullPolicy(),
		Resources:       instance.GetAppResources(),
		Args:            args,
		Env: []corev1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: v1.WebPort,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/api/readyz",
					Port:   intstr.FromInt(v1.WebPort),
					Scheme: "HTTPS",
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "tls-server",
				MountPath: v1.ServerCertificateMountPath,
				ReadOnly:  true,
			},
			{
				Name:      "tls-client",
				MountPath: v1.ClientCertificateMountPath,
				ReadOnly:  true,
			},
		},
	}
}

func newGrafanaContainerForCR(instance *appv1.VDICluster) corev1.Container {
	return corev1.Container{
		Name:            "grafana",
		Image:           "grafana/grafana",
		ImagePullPolicy: instance.GetAppPullPolicy(),
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: common.BoolPointer(true),
			RunAsUser:    common.Int64Ptr(1000),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "grafana-configs",
				MountPath: "/etc/grafana/provisioning/datasources",
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH",
				Value: "/etc/grafana/provisioning/datasources/dashboard.json",
			},
			{
				Name:  "GF_SERVER_ROOT_URL",
				Value: "http://localhost:3000/api/grafana",
			},
			{
				Name:  "GF_SERVER_SERVE_FROM_SUB_PATH",
				Value: "true",
			},
			{
				Name:  "GF_SERVER_ENABLE_GZIP",
				Value: "true",
			},
			{
				Name:  "GF_DEFAULT_INSTANCE_NAME",
				Value: "kvdi-grafana",
			},
			{
				Name:  "GF_AUTH_ORG_NAME",
				Value: "kvdi",
			},
			{
				Name:  "GF_AUTH_ORG_ROLE",
				Value: "Viewer",
			},
			{
				Name:  "GF_AUTH_DISABLE_LOGIN_FORM",
				Value: "true",
			},
			{
				Name:  "GF_AUTH_DISABLE_SIGNOUT_MENU",
				Value: "true",
			},
			{
				Name:  "GF_AUTH_ANONYMOUS_ENABLED",
				Value: "true",
			},
			{
				Name:  "GF_SECURITY_ALLOW_EMBEDDING",
				Value: "true",
			},
			{
				Name:  "GF_EXPLORE_ENABLED",
				Value: "false",
			},
		},
	}
}
