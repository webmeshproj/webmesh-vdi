package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DesktopTemplate is the Schema for the desktoptemplates API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=desktoptemplates,scope=Cluster
type DesktopTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DesktopTemplateSpec   `json:"spec,omitempty"`
	Status DesktopTemplateStatus `json:"status,omitempty"`
}

// DesktopInit represents the init system that the desktop container uses.
// +kubebuilder:validation:Enum=supervisord;systemd
type DesktopInit string

const (
	// InitSupervisord signals that the image uses supervisord.
	InitSupervisord = "supervisord"
	// InitSystemd signals that the image uses systemd.
	InitSystemd = "systemd"
)

// SocketType represents the type of service listening on the display socket
// in the container image.
// +kubebuilder:validation:Enum=xvnc;xpra
type SocketType string

const (
	// SocketXVNC signals that Xvnc is used for the display server.
	SocketXVNC SocketType = "xvnc"
	// SocketXPRA signals that Xpra is used for the display server.
	SocketXPRA SocketType = "xpra"
)

// DesktopTemplateSpec defines the desired state of DesktopTemplate
type DesktopTemplateSpec struct {
	// The docker repository and tag to use for desktops booted from this template.
	Image string `json:"image"`
	// The pull policy to use when pulling the container image.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Any pull secrets required for pulling the container image.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Additional environment variables to pass to containers booted from this template.
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Optionally map additional information about the user (and potentially extended further
	// in the future) into the environment of desktops booted from this template. The keys in the
	// map are the environment variable to set inside the desktop, and the values are go templates
	// or strings to set to the value. Currently the go templates are only passed a `Session` object
	// containing the information in the claims for the user that created the desktop. For more information
	// see the [JWTCLaims object](./metav1.md#JWTClaims) and corresponding go types.
	EnvTemplates map[string]string `json:"envTemplates,omitempty"`
	// Resource requirements to apply to desktops booted from this template.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Configuration options for the instances. These are highly dependant on using
	// the Dockerfiles (or close derivitives) provided in this repository.
	Config *DesktopConfig `json:"config,omitempty"`
	// Volume configurations for the instances. These can be used for mounting custom
	// volumes at arbitrary paths in desktops.
	VolumeConfig *DesktopVolumeConfig `json:"volumeConfig,omitempty"`
	// Arbitrary tags for displaying in the app UI.
	Tags map[string]string `json:"tags,omitempty"`
}

// DesktopConfig represents configurations for the template and desktops booted
// from it.
type DesktopConfig struct {
	// A service account to tie to desktops booted from this template.
	// TODO: This should really be per-desktop and by user-grants.
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// Extra system capabilities to add to desktops booted from this template.
	Capabilities []corev1.Capability `json:"capabilities,omitempty"`
	// AllowRoot will pass the ENABLE_ROOT envvar to the container. In the Dockerfiles
	// in this repository, this will add the user to the sudo group and ability to
	// sudo with no password.
	AllowRoot bool `json:"allowRoot,omitempty"`
	// The address the VNC server listens on inside the image. This defaults to the
	// UNIX socket /var/run/kvdi/display.sock. The kvdi-proxy sidecar will forward
	// websockify requests validated by mTLS to this socket.
	// Must be in the format of `tcp://{host}:{port}` or `unix://{path}`.
	SocketAddr string `json:"socketAddr,omitempty"`
	// Override the address of the PulseAudio server that the proxy will try to connect to
	// when serving audio. This defaults to what the ubuntu/arch desktop images are configured
	// to do during init.
	PulseServer string `json:"pulseServer,omitempty"`
	// The type of service listening on the configured socket. Can either be `xpra` or
	// `xvnc`. Currently `xpra` is used to serve "app profiles" and `xvnc` to serve full
	// desktops. Defaults to `xvnc`.
	SocketType SocketType `json:"socketType,omitempty"`
	// AllowFileTransfer will mount the user's home directory inside the kvdi-proxy image.
	// This enables the API endpoint for exploring, downloading, and uploading files to
	// desktop sessions booted from this template.
	AllowFileTransfer bool `json:"allowFileTransfer,omitempty"`
	// The image to use for the sidecar that proxies mTLS connections to the local
	// VNC server inside the Desktop. Defaults to the public kvdi-proxy image
	// matching the version of the currrently running manager.
	ProxyImage string `json:"proxyImage,omitempty"`
	// The type of init system inside the image, currently only supervisord and systemd
	// are supported. Defaults to `supervisord` (but depending on how much I like systemd
	// in this use case, that could change).
	Init DesktopInit `json:"init,omitempty"`
}

// DesktopVolumeConfig represents configurations for volumes attached to pods booted from
// a template.
type DesktopVolumeConfig struct {
	// Additional volumes to attach to pods booted from this template. To mount them there
	// must be cooresponding `volumeMounts` or `volumeDevices` specified.
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Volume mounts for the desktop container.
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volume devices for the desktop container.
	VolumeDevices []corev1.VolumeDevice `json:"volumeDevices,omitempty"`
}

// DesktopTemplateStatus defines the observed state of DesktopTemplate
type DesktopTemplateStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DesktopTemplateList contains a list of DesktopTemplate
type DesktopTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DesktopTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DesktopTemplate{}, &DesktopTemplateList{})
}
