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
	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

// DindIsEnabled returns true if dind is enabled for instances from this template.
func (t *Template) DindIsEnabled() bool {
	return t.Spec.DindConfig != nil
}

// GetDindImage returns the image to use for the dind sidecar.
func (t *Template) GetDindImage() string {
	if t.Spec.DindConfig != nil && t.Spec.DindConfig.Image != "" {
		return t.Spec.DindConfig.Image
	}
	return "docker:dind"
}

// GetDindPullPolicy returns the pull policy for the proxy container.
func (t *Template) GetDindPullPolicy() corev1.PullPolicy {
	if t.Spec.DindConfig != nil && t.Spec.DindConfig.ImagePullPolicy != "" {
		return t.Spec.DindConfig.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

// GetDindResources returns the resources for the dind container.
func (t *Template) GetDindResources() corev1.ResourceRequirements {
	if t.Spec.DindConfig != nil {
		return t.Spec.DindConfig.Resources
	}
	return corev1.ResourceRequirements{}
}

// GetDindVolumeMounts returns the volume mounts for the dind container.
func (t *Template) GetDindVolumeMounts() []corev1.VolumeMount {
	var mounts []corev1.VolumeMount
	if t.Spec.DindConfig != nil && len(t.Spec.DindConfig.VolumeMounts) > 0 {
		mounts = t.Spec.DindConfig.VolumeMounts
	} else {
		mounts = []corev1.VolumeMount{}
	}
	mounts = append(mounts, corev1.VolumeMount{
		Name:      v1.RunVolume,
		MountPath: v1.DesktopRunPath,
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      v1.DockerDataVolume,
		MountPath: v1.DockerDataPath,
	})
	return mounts
}

// GetDindVolumeDevices returns the volume devices for the dind container.
func (t *Template) GetDindVolumeDevices() []corev1.VolumeDevice {
	if t.Spec.DindConfig != nil && len(t.Spec.DindConfig.VolumeDevices) > 0 {
		return t.Spec.DindConfig.VolumeDevices
	}
	return nil
}

// GetDindContainer returns a dind sidecar to run for an instance, or nil if not configured
// on the template.
func (t *Template) GetDindContainer() corev1.Container {
	return corev1.Container{
		Name:            "dind",
		Image:           t.GetDindImage(),
		ImagePullPolicy: t.GetDindPullPolicy(),
		Resources:       t.GetDindResources(),
		VolumeMounts:    t.GetDindVolumeMounts(),
		VolumeDevices:   t.GetDindVolumeDevices(),
		SecurityContext: &corev1.SecurityContext{
			Privileged: &v1.True,
		},
	}
}
