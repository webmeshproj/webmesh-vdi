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
	"strconv"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

// GetDesktopContainer returns the container for the desktop.
func (t *Template) GetDesktopContainer(cluster *appv1.VDICluster, instance *Session, envSecret string) corev1.Container {
	c := corev1.Container{
		Name:            "desktop",
		Image:           t.GetDesktopImage(),
		ImagePullPolicy: t.GetDesktopPullPolicy(),
		VolumeMounts:    t.GetDesktopVolumeMounts(cluster, instance),
		VolumeDevices:   t.GetDesktopVolumeDevices(),
		SecurityContext: t.GetDesktopContainerSecurityContext(),
		Env:             t.GetDesktopEnvVars(instance),
		Lifecycle:       t.GetDesktopLifecycle(),
		Resources:       t.GetDesktopResources(),
	}
	if envSecret != "" {
		c.EnvFrom = []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: envSecret,
					},
				},
			},
		}
	}
	return c
}

// GetStaticEnvVars returns the environment variables configured in the template.
func (t *Template) GetStaticEnvVars() []corev1.EnvVar {
	if t.Spec.DesktopConfig != nil {
		return t.Spec.DesktopConfig.Env
	}
	return nil
}

// GetEnvTemplates returns the environment variable templates.
func (t *Template) GetEnvTemplates() map[string]string {
	if t.Spec.DesktopConfig != nil {
		return t.Spec.DesktopConfig.EnvTemplates
	}
	return nil
}

// GetDesktopVolumeDevices returns the additional volume devices to apply to the desktop container.
func (t *Template) GetDesktopVolumeDevices() []corev1.VolumeDevice {
	if t.Spec.DesktopConfig != nil && len(t.Spec.DesktopConfig.VolumeDevices) > 0 {
		return t.Spec.DesktopConfig.VolumeDevices
	}
	return nil
}

// GetInitSystem returns the init system used by the docker image in this template.
func (t *Template) GetInitSystem() DesktopInit {
	if t.Spec.DesktopConfig != nil && t.Spec.DesktopConfig.Init != "" {
		return t.Spec.DesktopConfig.Init
	}
	return InitSystemd
}

// RootEnabled returns true if desktops booted from the template should allow
// users to use sudo.
func (t *Template) RootEnabled() bool {
	if t.Spec.DesktopConfig != nil {
		return t.Spec.DesktopConfig.AllowRoot
	}
	return false
}

// GetDesktopImage returns the docker image to use for instances booted from
// this template.
func (t *Template) GetDesktopImage() string {
	if t.Spec.DesktopConfig != nil {
		return t.Spec.DesktopConfig.Image
	}
	return ""
}

// GetDesktopPullPolicy returns the image pull policy for this template.
func (t *Template) GetDesktopPullPolicy() corev1.PullPolicy {
	if t.Spec.DesktopConfig != nil && t.Spec.DesktopConfig.ImagePullPolicy != "" {
		return t.Spec.DesktopConfig.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

// GetDesktopResources returns the resource requirements for this instance.
func (t *Template) GetDesktopResources() corev1.ResourceRequirements {
	if t.Spec.DesktopConfig != nil {
		return t.Spec.DesktopConfig.Resources
	}
	return corev1.ResourceRequirements{}
}

// GetDesktopEnvVars returns the environment variables for a desktop pod.
func (t *Template) GetDesktopEnvVars(desktop *Session) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  v1.UserEnvVar,
			Value: desktop.GetUser(),
		},
		{
			Name:  v1.UIDEnvVar,
			Value: strconv.Itoa(int(v1.DefaultUser)),
		},
		{
			Name:  v1.HomeEnvVar,
			Value: fmt.Sprintf(v1.DesktopHomeFmt, desktop.GetUser()),
		},
	}
	if t.IsUNIXDisplaySocket() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  v1.VNCSockEnvVar,
			Value: t.GetDisplaySocketAddress(),
		})
	}
	if t.RootEnabled() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  v1.EnableRootEnvVar,
			Value: "true",
		})
	}
	if static := t.GetStaticEnvVars(); static != nil {
		envVars = append(envVars, static...)
	}
	return envVars
}

// GetDesktopContainerSecurityContext returns the container security context for
// pods booted from this template.
func (t *Template) GetDesktopContainerSecurityContext() *corev1.SecurityContext {
	capabilities := make([]corev1.Capability, 0)
	var privileged bool
	var user int64
	if t.GetInitSystem() == InitSystemd {
		// The method of using systemd-logind to trigger a systemd --user process
		// requires CAP_SYS_ADMIN. Specifically, SECCOMP spawning. There might
		// be other ways around this by just using system unit files for everything.
		capabilities = append(capabilities, "SYS_ADMIN")
		privileged = true
		user = 0
	} else {
		privileged = false
		user = v1.DefaultUser
	}
	if t.Spec.DesktopConfig != nil {
		capabilities = append(capabilities, t.Spec.DesktopConfig.Capabilities...)
	}
	return &corev1.SecurityContext{
		Privileged: &privileged,
		RunAsUser:  &user,
		Capabilities: &corev1.Capabilities{
			Add: capabilities,
		},
	}
}

// GetDesktopLifecycle returns the lifecycle actions for a desktop container booted from
// this template.
func (t *Template) GetDesktopLifecycle() *corev1.Lifecycle {
	if t.GetInitSystem() == InitSystemd {
		return &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"kill", "-s", "SIGRTMIN+3", "1"},
				},
			},
		}
	}
	return &corev1.Lifecycle{}
}
