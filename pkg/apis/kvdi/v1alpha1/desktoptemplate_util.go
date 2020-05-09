package v1alpha1

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tinyzimmer/kvdi/version"
	corev1 "k8s.io/api/core/v1"
)

// GetInitSystem returns the init system used by the docker image in this template.
func (t *DesktopTemplate) GetInitSystem() DesktopInit {
	if t.Spec.Config != nil && t.Spec.Config.Init != "" {
		return t.Spec.Config.Init
	}
	return InitSupervisord
}

// SoundEnabled returns true if the template supports virtual sound devices.
func (t *DesktopTemplate) SoundEnabled() bool {
	if t.Spec.Config != nil {
		return t.Spec.Config.EnableSound
	}
	return false
}

// RootEnabled returns true if desktops booted from the template should allow
// users to use sudo.
func (t *DesktopTemplate) RootEnabled() bool {
	if t.Spec.Config != nil {
		return t.Spec.Config.AllowRoot
	}
	return false
}

// GetNoVNCProxyImage returns the novnc-proxy image for the desktop instance.
func (t *DesktopTemplate) GetNoVNCProxyImage() string {
	if t.Spec.Config != nil && t.Spec.Config.ProxyImage != "" {
		return t.Spec.Config.ProxyImage
	}
	return fmt.Sprintf("quay.io/tinyzimmer/kvdi:novnc-proxy-%s", version.Version)
}

// GetDesktopImage returns the docker image to use for instances booted from
// this template.
func (t *DesktopTemplate) GetDesktopImage() string {
	if t.Spec.Image != "" {
		return t.Spec.Image
	}
	return fmt.Sprintf("quay.io/tinyzimmer/kvdi:desktop-%s", version.Version)
}

// GetDesktopPullPolicy returns the image pull policy for this template.
func (t *DesktopTemplate) GetDesktopPullPolicy() corev1.PullPolicy {
	if t.Spec.ImagePullPolicy != "" {
		return t.Spec.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

// GetDesktopPullSecrets returns the pull secrets for this instance.
func (t *DesktopTemplate) GetDesktopPullSecrets() []corev1.LocalObjectReference {
	return t.Spec.ImagePullSecrets
}

// GetDesktopResources returns the resource requirements for this instance.
func (t *DesktopTemplate) GetDesktopResources() corev1.ResourceRequirements {
	return t.Spec.Resources
}

// GetDesktopServiceAccount returns the service account for this instance.
// TODO: Should there be a default one?
func (t *DesktopTemplate) GetDesktopServiceAccount() string {
	if t.Spec.Config != nil {
		return t.Spec.Config.ServiceAccount
	}
	return ""
}

// GetVNCSocketAddr returns the VNC socket address to pass to the nonvnc-proxy.
func (t *DesktopTemplate) GetVNCSocketAddr() string {
	if t.Spec.Config != nil && t.Spec.Config.SocketAddr != "" {
		return t.Spec.Config.SocketAddr
	}
	return "unix:///var/run/kvdi/vnc.sock"
}

// GetDesktopEnvVars returns the environment variables for a desktop pod.
func (t *DesktopTemplate) GetDesktopEnvVars(desktop *Desktop) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "USER",
			Value: desktop.GetUser(),
		},
		{
			Name:  "VNC_SOCK_ADDR",
			Value: "/var/run/kvdi/vnc.sock",
		},
	}
	if t.RootEnabled() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ENABLE_ROOT",
			Value: "true",
		})
	}
	return envVars
}

// GetDesktopPodSecurityContext returns the security context for pods booted
// from this template.
func (t *DesktopTemplate) GetDesktopPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsNonRoot: &falseVal,
	}
}

// GetDesktopContainerSecurityContext returns the container security context for
// pods booted from this template.
func (t *DesktopTemplate) GetDesktopContainerSecurityContext() *corev1.SecurityContext {
	capabilities := make([]corev1.Capability, 0)
	if t.GetInitSystem() == InitSystemd {
		// The method of using systemd-logind to trigger a systemd --user process
		// requires CAP_SYS_ADMIN. Specifically, SECCOMP spawing. There might
		// be other ways around this by just using system unit files for everything.
		capabilities = append(capabilities, "SYS_ADMIN")
	}
	if t.Spec.Config != nil {
		capabilities = append(capabilities, t.Spec.Config.Capabilities...)
	}
	return &corev1.SecurityContext{
		Privileged: &trueVal,
		Capabilities: &corev1.Capabilities{
			Add: capabilities,
		},
	}
}

// GetDesktopVolumes returns the volumes to mount to desktop pods.
func (t *DesktopTemplate) GetDesktopVolumes(cluster *VDICluster, desktop *Desktop) []corev1.Volume {
	// Common volumes all containers will need.
	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: "run",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: "run-lock",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "vnc-sock",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "shm",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/dev/shm",
				},
			},
		},
		{
			Name: "tls",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: desktop.GetName(),
				},
			},
		},
	}

	// A PVC claim for the user if specified, otherwise use an EmptyDir.
	if cluster.GetUserdataVolumeSpec() != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "home",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: cluster.GetUserdataVolumeName(desktop.GetUser()),
				},
			},
		})
	} else {
		volumes = append(volumes, corev1.Volume{
			Name: "home",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// If systemd we need to add a few more temp filesystems and bind mount
	// /sys/fs/cgroup.
	if t.GetInitSystem() == InitSystemd {
		volumes = append(volumes, []corev1.Volume{
			{
				Name: "cgroup",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/sys/fs/cgroup",
					},
				},
			},
		}...)
	}

	return volumes
}

// GetDesktopVolumeMounts returns the volume mounts for the main desktop container.
func (t *DesktopTemplate) GetDesktopVolumeMounts(cluster *VDICluster, desktop *Desktop) []corev1.VolumeMount {

	mounts := []corev1.VolumeMount{
		{
			Name:      "tmp",
			MountPath: "/tmp",
		},
		{
			Name:      "run",
			MountPath: "/run",
		},
		{
			Name:      "run-lock",
			MountPath: "/run/lock",
		},
		{
			Name:      "vnc-sock",
			MountPath: filepath.Dir(strings.TrimPrefix(strings.TrimPrefix(t.GetVNCSocketAddr(), "unix://"), "tcp://")),
		},
		{
			Name:      "shm",
			MountPath: "/dev/shm",
		},
		{
			Name:      "home",
			MountPath: fmt.Sprintf("/home/%s", desktop.GetUser()),
		},
	}
	if t.GetInitSystem() == InitSystemd {
		mounts = append(mounts, []corev1.VolumeMount{
			{
				Name:      "cgroup",
				MountPath: "/sys/fs/cgroup",
			},
		}...)
	}
	return mounts
}

// GetDesktopProxyContainer returns the configuration for the novnc-proxy sidecar.
func (t *DesktopTemplate) GetDesktopProxyContainer() corev1.Container {
	return corev1.Container{
		Name:            "novnc-proxy",
		Image:           t.GetNoVNCProxyImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args:            []string{"--vnc-addr", t.GetVNCSocketAddr()},
		Ports: []corev1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: WebPort,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "run",
				MountPath: "/run",
			},
			{
				Name:      "run-lock",
				MountPath: "/run/lock",
			},
			{
				Name:      "tls",
				MountPath: ServerCertificateMountPath,
				ReadOnly:  true,
			},
			{
				Name:      "vnc-sock",
				MountPath: filepath.Dir(strings.TrimPrefix(strings.TrimPrefix(t.GetVNCSocketAddr(), "unix://"), "tcp://")),
			},
		},
	}
}

// GetLifecycle returns the lifecycle actions for a desktop container booted from
// this template.
func (t *DesktopTemplate) GetLifecycle() *corev1.Lifecycle {
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
