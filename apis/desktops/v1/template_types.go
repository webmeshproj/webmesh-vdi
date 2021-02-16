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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DesktopInit represents the init system that the desktop container uses.
// +kubebuilder:validation:Enum=supervisord;systemd
type DesktopInit string

const (
	// InitSupervisord signals that the image uses supervisord.
	InitSupervisord = "supervisord"
	// InitSystemd signals that the image uses systemd.
	InitSystemd = "systemd"
)

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	// Any pull secrets required for pulling the container image.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Additional volumes to attach to pods booted from this template. To mount them there
	// must be cooresponding `volumeMounts` or `volumeDevices` specified.
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Configuration options for the instances. These are highly dependant on using
	// the Dockerfiles (or close derivitives) provided in this repository.
	DesktopConfig *DesktopConfig `json:"desktop,omitempty"`
	// Configurations for the display proxy.
	ProxyConfig *ProxyConfig `json:"proxy,omitempty"`
	// Docker-in-docker configurations for running a dind sidecar along with desktop instances.
	DindConfig *DockerInDockerConfig `json:"dind,omitempty"`
	// QEMU configurations for this template. When defined, VMs are used instead of containers
	// for desktop sessions. This object is mututally exclusive with `desktop` and will take
	// precedence when defined.
	QEMUConfig *QEMUConfig `json:"qemu,omitempty"`
	// Arbitrary tags for displaying in the app UI.
	Tags map[string]string `json:"tags,omitempty"`
}

// DesktopConfig represents configurations for the template and desktops booted
// from it.
type DesktopConfig struct {
	// The docker repository and tag to use for desktops booted from this template.
	Image string `json:"image,omitempty"`
	// The pull policy to use when pulling the container image.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Resource requirements to apply to desktops booted from this template.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Additional environment variables to pass to containers booted from this template.
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Optionally map additional information about the user (and potentially extended further
	// in the future) into the environment of desktops booted from this template. The keys in the
	// map are the environment variable to set inside the desktop, and the values are go templates
	// or strings to set to the value. Currently the go templates are only passed a `Session` object
	// containing the information in the claims for the user that created the desktop. For more information
	// see the [JWTCaims object](https://github.com/tinyzimmer/kvdi/blob/main/pkg/types/auth_types.go#L79)
	// and corresponding go types.
	EnvTemplates map[string]string `json:"envTemplates,omitempty"`
	// Volume mounts for the desktop container.
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volume devices for the desktop container.
	VolumeDevices []corev1.VolumeDevice `json:"volumeDevices,omitempty"`
	// Extra system capabilities to add to desktops booted from this template.
	Capabilities []corev1.Capability `json:"capabilities,omitempty"`
	// AllowRoot will pass the ENABLE_ROOT envvar to the container. In the Dockerfiles
	// in this repository, this will add the user to the sudo group and ability to
	// sudo with no password.
	AllowRoot bool `json:"allowRoot,omitempty"`
	// The type of init system inside the image, currently only `supervisord` and `systemd`
	// are supported. Defaults to `systemd`. `systemd` containers are run privileged and
	// downgrading to the desktop user must be done within the image's init process. `supervisord`
	// containers are run with minimal capabilities and directly as the desktop user.
	Init DesktopInit `json:"init,omitempty"`
}

// ProxyConfig represents configurations for the display/audio proxy.
type ProxyConfig struct {
	// The image to use for the sidecar that proxies mTLS connections to the local
	// VNC server inside the Desktop. Defaults to the public kvdi-proxy image
	// matching the version of the currrently running manager.
	Image string `json:"image,omitempty"`
	// The pull policy to use when pulling the container image.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// AllowFileTransfer will mount the user's home directory inside the kvdi-proxy image.
	// This enables the API endpoint for exploring, downloading, and uploading files to
	// desktop sessions booted from this template. When using a `qemu` configuration with
	// SPICE, file upload is enabled by default.
	AllowFileTransfer bool `json:"allowFileTransfer,omitempty"`
	// The address the display server listens on inside the image. This defaults to the
	// UNIX socket `/var/run/kvdi/display.sock`. The kvdi-proxy sidecar will forward
	// websockify requests validated by mTLS to this socket. Must be in the format of
	// `tcp://{host}:{port}` or `unix://{path}`. This will usually be a VNC server unless
	// using a `qemu` configuration with SPICE. If using custom init scripts inside your
	// containers, this value is set to the `DISPLAY_SOCK_ADDR` environment variable.
	SocketAddr string `json:"socketAddr,omitempty"`
	// Override the address of the PulseAudio server that the proxy will try to connect to
	// when serving audio. This defaults to what the ubuntu/arch desktop images are configured
	// to do during init, which is to place a socket in the user's run directory. The value is
	// assumed to be a unix socket.
	PulseServer string `json:"pulseServer,omitempty"`
	// Resource restraints to place on the proxy sidecar.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// DockerInDockerConfig is a configuration for mounting a DinD sidecar with desktops
// booted from the template. This will provide ephemeral docker daemons and storage
// to sessions.
type DockerInDockerConfig struct {
	// The image to use for the dind sidecar. Defaults to `docker:dind`.
	Image string `json:"image,omitempty"`
	// The pull policy to use when pulling the container image.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Resource restraints to place on the dind sidecar.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Volume mounts for the dind container.
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volume devices for the dind container.
	VolumeDevices []corev1.VolumeDevice `json:"volumeDevices,omitempty"`
}

// QEMUConfig represents configurations for running a qemu virtual machine for instances
// booted from this template.
type QEMUConfig struct {
	// The container image bundling the disks for this template.
	DiskImage string `json:"diskImage,omitempty"`
	// The pull policy to use when pulling the disk image.
	DiskImagePullPolicy corev1.PullPolicy `json:"diskImagePullPolicy,omitempty"`
	// Set to true to use the image-populator CSI to mount the disk images to a qemu container.
	// You must have the [image-populator](https://github.com/kubernetes-csi/csi-driver-image-populator)
	// driver installed. Defaults to copying the contents out of the disk image via an init
	// container. This is experimental and not really tested.
	UseCSI bool `json:"useCSI,omitempty"`
	// The container image containing the QEMU utilities to use to launch the VM.
	// Defaults to `ghcr.io/tinyzimmer/kvdi:qemu-latest`.
	QEMUImage string `json:"qemuImage,omitempty"`
	// The pull policy to use when pulling the QEMU image.
	QEMUImagePullPolicy corev1.PullPolicy `json:"qemuImagePullPolicy,omitempty"`
	// Resource requirements to place on the qemu runner instance.
	QEMUResources corev1.ResourceRequirements `json:"qemuResources,omitempty"`
	// The path to the boot volume inside the disk image. Defaults to `/disk/boot.img`.
	DiskPath string `json:"diskPath,omitempty"`
	// The path to a pre-built cloud init image to use when booting the VM inside the disk
	// image. Defaults to an auto-generated one at runtime.
	CloudInitPath string `json:"cloudInitPath,omitempty"`
	// The number of vCPUs to assign the virtual machine. Defaults to 1.
	CPUs int `json:"cpus,omitempty"`
	// The amount of memory to assign the virtual machine (in MB). Defaults to 1024.
	Memory int `json:"memory,omitempty"`
	// Set to true to use the SPICE protocol when proxying the display. If using custom qemu runners,
	// this sets the `SPICE_DISPLAY` environment variable to `true`. The runners provided by this
	// repository will tell qemu to set up a SPICE server at `proxy.socketAddr`. The default is to use
	// VNC. This value is also used by the UI to determine which protocol to expect from a display connection.
	SPICE bool `json:"spice,omitempty"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=templates,scope=Cluster
//+kubebuilder:subresource:status

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
