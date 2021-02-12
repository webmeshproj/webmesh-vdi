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
	"path"
	"path/filepath"
	"strconv"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/version"

	corev1 "k8s.io/api/core/v1"
)

// GetInitContainers returns any init containers required to run before the desktop launches.
func (t *Template) GetInitContainers() []corev1.Container {
	containers := make([]corev1.Container, 0)
	if t.DindIsEnabled() {
		containers = append(containers, corev1.Container{
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
		})
	}
	return containers
}

// GetStaticEnvVars returns the environment variables configured in the template.
func (t *Template) GetStaticEnvVars() []corev1.EnvVar {
	if t.Spec.Config != nil {
		return t.Spec.Config.Env
	}
	return nil
}

// GetEnvTemplates returns the environment variable templates.
func (t *Template) GetEnvTemplates() map[string]string {
	if t.Spec.Config != nil {
		return t.Spec.Config.EnvTemplates
	}
	return nil
}

// GetPulseServer returns the pulse server to give to the proxy for handling audio streams.
func (t *Template) GetPulseServer() string {
	if t.Spec.ProxyConfig != nil && t.Spec.ProxyConfig.PulseServer != "" {
		return strings.TrimPrefix(t.Spec.ProxyConfig.PulseServer, "unix://")
	}
	return fmt.Sprintf("/run/user/%d/pulse/native", v1.DefaultUser)
}

// GetVolumes returns the additional volumes to apply to a pod.
func (t *Template) GetVolumes() []corev1.Volume {
	return t.Spec.Volumes
}

// GetVolumeMounts returns the additional volume mounts to apply to the desktop container.
func (t *Template) GetVolumeMounts() []corev1.VolumeMount {
	if t.Spec.Config != nil && len(t.Spec.Config.VolumeMounts) > 0 {
		return t.Spec.Config.VolumeMounts
	}
	return nil
}

// GetVolumeDevices returns the additional volume devices to apply to the desktop container.
func (t *Template) GetVolumeDevices() []corev1.VolumeDevice {
	if t.Spec.Config != nil && len(t.Spec.Config.VolumeDevices) > 0 {
		return t.Spec.Config.VolumeDevices
	}
	return nil
}

// GetInitSystem returns the init system used by the docker image in this template.
func (t *Template) GetInitSystem() DesktopInit {
	if t.Spec.Config != nil && t.Spec.Config.Init != "" {
		return t.Spec.Config.Init
	}
	return InitSystemd
}

// RootEnabled returns true if desktops booted from the template should allow
// users to use sudo.
func (t *Template) RootEnabled() bool {
	if t.Spec.Config != nil {
		return t.Spec.Config.AllowRoot
	}
	return false
}

// FileTransferEnabled returns true if desktops booted from the template should
// allow file transfer.
func (t *Template) FileTransferEnabled() bool {
	if t.Spec.ProxyConfig != nil {
		return t.Spec.ProxyConfig.AllowFileTransfer
	}
	return false
}

// GetKVDIVNCProxyImage returns the kvdi-proxy image for the desktop instance.
func (t *Template) GetKVDIVNCProxyImage() string {
	if t.Spec.ProxyConfig != nil && t.Spec.ProxyConfig.Image != "" {
		return t.Spec.ProxyConfig.Image
	}
	return fmt.Sprintf("ghcr.io/tinyzimmer/kvdi:kvdi-proxy-%s", version.Version)
}

// GetDesktopImage returns the docker image to use for instances booted from
// this template.
func (t *Template) GetDesktopImage() string {
	if t.Spec.Config != nil {
		return t.Spec.Config.Image
	}
	return ""
}

// GetDesktopPullPolicy returns the image pull policy for this template.
func (t *Template) GetDesktopPullPolicy() corev1.PullPolicy {
	if t.Spec.Config != nil && t.Spec.Config.ImagePullPolicy != "" {
		return t.Spec.Config.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

// GetDesktopPullSecrets returns the pull secrets for this instance.
func (t *Template) GetDesktopPullSecrets() []corev1.LocalObjectReference {
	return t.Spec.ImagePullSecrets
}

// GetDesktopResources returns the resource requirements for this instance.
func (t *Template) GetDesktopResources() corev1.ResourceRequirements {
	if t.Spec.Config != nil {
		return t.Spec.Config.Resources
	}
	return corev1.ResourceRequirements{}
}

// IsTCPDisplaySocket returns true if the VNC server is listening on a TCP socket.
func (t *Template) IsTCPDisplaySocket() bool {
	return strings.HasPrefix(t.GetDisplaySocketURI(), "tcp://")
}

// IsUNIXDisplaySocket returns true if the VNC server is listening on a UNIX socket.
func (t *Template) IsUNIXDisplaySocket() bool {
	return strings.HasPrefix(t.GetDisplaySocketURI(), "unix://")
}

// GetDisplaySocketAddress returns just the address portion of the display socket URI.
func (t *Template) GetDisplaySocketAddress() string {
	return strings.TrimPrefix(strings.TrimPrefix(t.GetDisplaySocketURI(), "unix://"), "tcp://")
}

// GetDisplaySocketURI returns the display socket URI to pass to the nonvnc-proxy.
func (t *Template) GetDisplaySocketURI() string {
	if t.Spec.ProxyConfig != nil && t.Spec.ProxyConfig.SocketAddr != "" {
		return t.Spec.ProxyConfig.SocketAddr
	}
	return v1.DefaultDisplaySocketAddr
}

// GetDesktopEnvVars returns the environment variables for a desktop pod.
func (t *Template) GetDesktopEnvVars(desktop *Session) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  v1.UserEnvVar,
			Value: desktop.GetUser(),
		},
		{
			Name:  "UID",
			Value: strconv.Itoa(int(v1.DefaultUser)), // TODO: Better here than in the images, but still needs refactoring
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

// GetDesktopPodSecurityContext returns the security context for pods booted
// from this template.
func (t *Template) GetDesktopPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsNonRoot: &v1.False,
	}
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
	if t.Spec.Config != nil {
		capabilities = append(capabilities, t.Spec.Config.Capabilities...)
	}
	return &corev1.SecurityContext{
		Privileged: &privileged,
		RunAsUser:  &user,
		Capabilities: &corev1.Capabilities{
			Add: capabilities,
		},
	}
}

// NeedsDedicatedPulseVolume returns true if the location of the pulse socket is not
// covered by any of the existing mounts.
func (t *Template) NeedsDedicatedPulseVolume() bool {
	if t.IsUNIXDisplaySocket() {
		if filepath.Dir(t.GetDisplaySocketAddress()) == filepath.Dir(t.GetPulseServer()) {
			return false
		}
	}
	if t.Spec.Config != nil && len(t.Spec.Config.VolumeMounts) > 0 {
		for _, mount := range t.Spec.Config.VolumeMounts {
			if strings.HasPrefix(t.GetPulseServer(), mount.MountPath) {
				return false
			}
		}
	}
	for _, path := range []string{v1.DesktopTmpPath, v1.DesktopRunPath, "/home"} {
		if strings.HasPrefix(t.GetPulseServer(), path) {
			return false
		}
	}
	return true
}

// GetProxyPullPolicy returns the pull policy for the proxy container.
func (t *Template) GetProxyPullPolicy() corev1.PullPolicy {
	if t.Spec.ProxyConfig != nil && t.Spec.ProxyConfig.ImagePullPolicy != "" {
		return t.Spec.ProxyConfig.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

// GetProxyResources returns the resources for the proxy container.
func (t *Template) GetProxyResources() corev1.ResourceRequirements {
	if t.Spec.ProxyConfig != nil {
		return t.Spec.ProxyConfig.Resources
	}
	return corev1.ResourceRequirements{}
}

// GetDesktopProxyContainer returns the configuration for the kvdi-proxy sidecar.
func (t *Template) GetDesktopProxyContainer() corev1.Container {
	proxyVolMounts := []corev1.VolumeMount{
		{
			Name:      v1.TmpVolume,
			MountPath: v1.DesktopTmpPath,
		},
		{
			Name:      v1.RunVolume,
			MountPath: v1.DesktopRunPath,
		},
		{
			Name:      v1.RunLockVolume,
			MountPath: v1.DesktopRunLockPath,
		},
		{
			Name:      v1.TLSVolume,
			MountPath: v1.ServerCertificateMountPath,
			ReadOnly:  true,
		},
	}
	if t.IsUNIXDisplaySocket() && !strings.HasPrefix(path.Dir(t.GetDisplaySocketAddress()), v1.DesktopTmpPath) {
		proxyVolMounts = append(proxyVolMounts, corev1.VolumeMount{
			Name:      v1.VNCSockVolume,
			MountPath: filepath.Dir(t.GetDisplaySocketAddress()),
		})
	}
	if t.NeedsDedicatedPulseVolume() {
		proxyVolMounts = append(proxyVolMounts, corev1.VolumeMount{
			Name:      v1.PulseSockVolume,
			MountPath: filepath.Dir(t.GetPulseServer()),
		})
	}
	if t.FileTransferEnabled() {
		proxyVolMounts = append(proxyVolMounts, corev1.VolumeMount{
			Name:      v1.HomeVolume,
			MountPath: v1.DesktopHomeMntPath,
		})
	}
	return corev1.Container{
		Name:            "kvdi-proxy",
		Image:           t.GetKVDIVNCProxyImage(),
		ImagePullPolicy: t.GetProxyPullPolicy(),
		Args: []string{
			"--vnc-addr", t.GetDisplaySocketURI(),
			"--user-id", strconv.Itoa(int(v1.DefaultUser)),
			"--pulse-server", t.GetPulseServer(),
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: v1.WebPort,
			},
		},
		VolumeMounts: proxyVolMounts,
		Resources:    t.GetProxyResources(),
	}
}

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

// GetLifecycle returns the lifecycle actions for a desktop container booted from
// this template.
func (t *Template) GetLifecycle() *corev1.Lifecycle {
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
