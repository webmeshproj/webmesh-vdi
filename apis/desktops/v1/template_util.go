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

package v1

import (
	"fmt"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// ToPodSpec computes a `corev1.PodSpec` from this template given a parent cluster, user session, and optional
// environment variable secret name.
func (t *Template) ToPodSpec(cluster *appv1.VDICluster, instance *Session, envSecret string) corev1.PodSpec {
	return corev1.PodSpec{
		Hostname:           instance.GetName(),
		Subdomain:          instance.GetName(),
		ServiceAccountName: instance.GetServiceAccount(),
		SecurityContext:    t.GetPodSecurityContext(),
		Volumes:            t.GetVolumes(cluster, instance),
		ImagePullSecrets:   t.GetPullSecrets(),
		InitContainers:     t.GetInitContainers(),
		Containers:         t.GetContainers(cluster, instance, envSecret),
	}
}

// GetContainers returns the containers for a given Session.
func (t *Template) GetContainers(cluster *appv1.VDICluster, instance *Session, envSecret string) []corev1.Container {
	containers := []corev1.Container{t.GetDesktopProxyContainer()}
	if t.IsQEMUTemplate() {
		containers = append(containers, t.GetQEMUContainer(cluster, instance))
	} else {
		containers = append(containers, t.GetDesktopContainer(cluster, instance, envSecret))
	}
	if t.DindIsEnabled() {
		containers = append(containers, t.GetDindContainer())
	}
	return containers
}

// GetInitContainers returns any init containers required to run before the desktop launches.
func (t *Template) GetInitContainers() []corev1.Container {
	if t.IsQEMUTemplate() && !t.QEMUUseCSI() {
		cmd := fmt.Sprintf("cp %s %s && chmod 666 %s", t.GetQEMUDiskPath(), v1.QEMUNonCSIBootImagePath, v1.QEMUNonCSIBootImagePath)
		if cloudInit := t.GetQEMUCloudInitPath(); cloudInit != "" {
			cmd += fmt.Sprintf(" && cp %s %s && chmod 666 %s", cloudInit, v1.QEMUNonCSICloudImagePath, v1.QEMUNonCSICloudImagePath)
		}
		return []corev1.Container{
			{
				Name:            "qemu-kvm-init",
				Image:           t.GetQEMUDiskImage(),
				ImagePullPolicy: t.GetQEMUDiskImagePullPolicy(),
				Command:         []string{"/bin/sh", "-c", cmd},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      v1.RunVolume,
						MountPath: v1.DesktopRunPath,
					},
				},
			},
		}
	}
	if t.DindIsEnabled() {
		return []corev1.Container{
			{
				Name:            "dind-init",
				Image:           t.GetDindImage(),
				ImagePullPolicy: t.GetDindPullPolicy(),
				Command:         []string{"/bin/sh", "-c", fmt.Sprintf("cp -r /usr/local/bin/* %s", v1.DockerBinPath)},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      v1.DockerBinVolume,
						MountPath: v1.DockerBinPath,
					},
				},
			},
		}
	}
	return nil
}

// GetPullSecrets returns the pull secrets for this instance.
func (t *Template) GetPullSecrets() []corev1.LocalObjectReference {
	return t.Spec.ImagePullSecrets
}

// GetPodSecurityContext returns the security context for pods booted
// from this template.
func (t *Template) GetPodSecurityContext() *corev1.PodSecurityContext {
	if t.DindIsEnabled() || t.GetInitSystem() == InitSystemd {
		return &corev1.PodSecurityContext{
			RunAsNonRoot: &v1.False,
		}
	}
	return &corev1.PodSecurityContext{
		RunAsNonRoot: &v1.True,
		RunAsUser:    &v1.DefaultUser,
	}
}
