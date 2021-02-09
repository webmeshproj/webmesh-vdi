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
	"k8s.io/apimachinery/pkg/api/resource"
)

// GetStaticEnvVars returns the environment variables configured in the template.
func (t *Template) GetStaticEnvVars() []corev1.EnvVar { return t.Spec.Env }

// GetEnvTemplates returns the environment variable templates.
func (t *Template) GetEnvTemplates() map[string]string { return t.Spec.EnvTemplates }

// GetPulseServer returns the pulse server to give to the proxy for handling audio streams.
func (t *Template) GetPulseServer() string {
	if t.Spec.Config != nil && t.Spec.Config.PulseServer != "" {
		return strings.TrimPrefix(t.Spec.Config.PulseServer, "unix://")
	}
	return fmt.Sprintf("/run/user/%d/pulse/native", v1.DefaultUser)
}

// GetVolumes returns the additional volumes to apply to a pod.
func (t *Template) GetVolumes() []corev1.Volume {
	if t.Spec.VolumeConfig != nil && t.Spec.VolumeConfig.Volumes != nil {
		return t.Spec.VolumeConfig.Volumes
	}
	return nil
}

// GetVolumeMounts returns the additional volume mounts to apply to the desktop container.
func (t *Template) GetVolumeMounts() []corev1.VolumeMount {
	if t.Spec.VolumeConfig != nil && t.Spec.VolumeConfig.VolumeMounts != nil {
		return t.Spec.VolumeConfig.VolumeMounts
	}
	return nil
}

// GetVolumeDevices returns the additional volume devices to apply to the desktop container.
func (t *Template) GetVolumeDevices() []corev1.VolumeDevice {
	if t.Spec.VolumeConfig != nil && t.Spec.VolumeConfig.VolumeDevices != nil {
		return t.Spec.VolumeConfig.VolumeDevices
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
	if t.Spec.Config != nil {
		return t.Spec.Config.AllowFileTransfer
	}
	return false
}

// GetKVDIVNCProxyImage returns the kvdi-proxy image for the desktop instance.
func (t *Template) GetKVDIVNCProxyImage() string {
	if t.Spec.Config != nil && t.Spec.Config.ProxyImage != "" {
		return t.Spec.Config.ProxyImage
	}
	return fmt.Sprintf("ghcr.io/tinyzimmer/kvdi:kvdi-proxy-%s", version.Version)
}

// GetDesktopImage returns the docker image to use for instances booted from
// this template.
func (t *Template) GetDesktopImage() string {
	return t.Spec.Image
}

// GetDesktopPullPolicy returns the image pull policy for this template.
func (t *Template) GetDesktopPullPolicy() corev1.PullPolicy {
	if t.Spec.ImagePullPolicy != "" {
		return t.Spec.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

// GetDesktopPullSecrets returns the pull secrets for this instance.
func (t *Template) GetDesktopPullSecrets() []corev1.LocalObjectReference {
	return t.Spec.ImagePullSecrets
}

// GetDesktopResources returns the resource requirements for this instance.
func (t *Template) GetDesktopResources() corev1.ResourceRequirements {
	return t.Spec.Resources
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
	if t.Spec.Config != nil && t.Spec.Config.SocketAddr != "" {
		return t.Spec.Config.SocketAddr
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

// Volume names
var (
	TmpVolume       = "tmp"
	RunVolume       = "run"
	ShmVolume       = "shm"
	TLSVolume       = "tls"
	HomeVolume      = "home"
	CgroupsVolume   = "cgroups"
	RunLockVolume   = "run-lock"
	VNCSockVolume   = "vnc-sock"
	PulseSockVolume = "pulse-sock"
)

// NeedsDedicatedPulseVolume returns true if the location of the pulse socket is not
// covered by any of the existing mounts.
func (t *Template) NeedsDedicatedPulseVolume() bool {
	if t.IsUNIXDisplaySocket() {
		if filepath.Dir(t.GetDisplaySocketAddress()) == filepath.Dir(t.GetPulseServer()) {
			return false
		}
	}
	if t.Spec.VolumeConfig != nil && len(t.Spec.VolumeConfig.VolumeMounts) > 0 {
		for _, mount := range t.Spec.VolumeConfig.VolumeMounts {
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

// GetDesktopProxyContainer returns the configuration for the kvdi-proxy sidecar.
func (t *Template) GetDesktopProxyContainer() corev1.Container {
	proxyVolMounts := []corev1.VolumeMount{
		{
			Name:      TmpVolume,
			MountPath: v1.DesktopTmpPath,
		},
		{
			Name:      RunVolume,
			MountPath: v1.DesktopRunPath,
		},
		{
			Name:      RunLockVolume,
			MountPath: v1.DesktopRunLockPath,
		},
		{
			Name:      TLSVolume,
			MountPath: v1.ServerCertificateMountPath,
			ReadOnly:  true,
		},
	}
	if t.IsUNIXDisplaySocket() && !strings.HasPrefix(path.Dir(t.GetDisplaySocketAddress()), v1.DesktopTmpPath) {
		proxyVolMounts = append(proxyVolMounts, corev1.VolumeMount{
			Name:      VNCSockVolume,
			MountPath: filepath.Dir(t.GetDisplaySocketAddress()),
		})
	}
	if t.NeedsDedicatedPulseVolume() {
		proxyVolMounts = append(proxyVolMounts, corev1.VolumeMount{
			Name:      PulseSockVolume,
			MountPath: filepath.Dir(t.GetPulseServer()),
		})
	}
	if t.FileTransferEnabled() {
		proxyVolMounts = append(proxyVolMounts, corev1.VolumeMount{
			Name:      HomeVolume,
			MountPath: v1.DesktopHomeMntPath,
		})
	}
	return corev1.Container{
		Name:            "kvdi-proxy",
		Image:           t.GetKVDIVNCProxyImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
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
		// TODO: Make these configurable
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
			// We need to be able to burst pretty high if the user wants to
			// download a large directory. An admin should be the one to determine
			// how many resources a user can use at any given time. This would also have
			// the benefit of limiting network traffic.
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
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
