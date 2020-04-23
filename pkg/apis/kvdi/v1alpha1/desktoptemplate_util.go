package v1alpha1

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/version"
	corev1 "k8s.io/api/core/v1"
)

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
	if t.Spec.Config == nil {
		return envVars
	}
	if t.Spec.Config.AllowRoot {
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
	// var capabilities []corev1.Capability
	// if t.Spec.Config != nil {
	// 	capabilities = t.Spec.Config.Capabilities
	// }
	return &corev1.SecurityContext{
		Privileged: &trueVal,
		// Capabilities: &corev1.Capabilities{
		// 	Drop: []corev1.Capability{"ALL"},
		// 	Add:  capabilities,
		// },
	}
}

// GetDesktopVolumes returns the volumes to mount to desktop pods.
// TODO: Persistent for users can be added here.
func (t *DesktopTemplate) GetDesktopVolumes(cluster *VDICluster, desktop *Desktop) []corev1.Volume {
	volumes := []corev1.Volume{
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
	if t.Spec.Config == nil {
		return volumes
	}
	if t.Spec.Config.EnableSound {
		volumes = append(volumes, corev1.Volume{
			Name: "sound",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/dev/snd",
				},
			},
		})
	}
	return volumes
}

// GetDesktopVolumeMounts returns the volume mounts for the main desktop container.
func (t *DesktopTemplate) GetDesktopVolumeMounts(cluster *VDICluster, desktop *Desktop) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      "vnc-sock",
			MountPath: "/var/run/kvdi",
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
	if t.Spec.Config == nil {
		return mounts
	}
	if t.Spec.Config.EnableSound {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "sound",
			MountPath: "/dev/snd",
		})
	}
	return mounts
}

// GetDesktopProxyContainer returns the configuration for the novnc-proxy sidecar.
func (t *DesktopTemplate) GetDesktopProxyContainer(proxyImg string) corev1.Container {
	return corev1.Container{
		Name:            "novnc-proxy",
		Image:           proxyImg,
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
				Name:      "tls",
				MountPath: ServerCertificateMountPath,
				ReadOnly:  true,
			},
			{
				Name:      "vnc-sock",
				MountPath: "/var/run/kvdi",
			},
		},
	}
}
