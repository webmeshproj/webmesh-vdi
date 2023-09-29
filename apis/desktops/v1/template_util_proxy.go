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

	corev1 "k8s.io/api/core/v1"

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/version"
)

// FileTransferEnabled returns true if desktops booted from the template should
// allow file transfer.
func (t *Template) FileTransferEnabled() bool {
	if t.Spec.ProxyConfig != nil {
		return t.Spec.ProxyConfig.AllowFileTransfer
	}
	return false
}

// GetPulseServer returns the pulse server to give to the proxy for handling audio streams.
func (t *Template) GetPulseServer() string {
	if t.Spec.ProxyConfig != nil && t.Spec.ProxyConfig.PulseServer != "" {
		return strings.TrimPrefix(t.Spec.ProxyConfig.PulseServer, "unix://")
	}
	return fmt.Sprintf("/run/user/%d/pulse/native", v1.DefaultUser)
}

// GetKVDIVNCProxyImage returns the kvdi-proxy image for the desktop instance.
func (t *Template) GetKVDIVNCProxyImage() string {
	if t.Spec.ProxyConfig != nil && t.Spec.ProxyConfig.Image != "" {
		return t.Spec.ProxyConfig.Image
	}
	return fmt.Sprintf("ghcr.io/webmeshproj/vdi-proxy:%s", version.Version)
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

// NeedsDedicatedPulseVolume returns true if the location of the pulse socket is not
// covered by any of the existing mounts.
func (t *Template) NeedsDedicatedPulseVolume() bool {
	if t.IsUNIXDisplaySocket() {
		if filepath.Dir(t.GetDisplaySocketAddress()) == filepath.Dir(t.GetPulseServer()) {
			return false
		}
	}
	if t.Spec.DesktopConfig != nil && len(t.Spec.DesktopConfig.VolumeMounts) > 0 {
		for _, mount := range t.Spec.DesktopConfig.VolumeMounts {
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
			Name:      t.GetTmpVolume(),
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
	c := corev1.Container{
		Name:            "kvdi-proxy",
		Image:           t.GetKVDIVNCProxyImage(),
		ImagePullPolicy: t.GetProxyPullPolicy(),
		Args: []string{
			"--display-addr", t.GetDisplaySocketURI(),
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

	return c
}
