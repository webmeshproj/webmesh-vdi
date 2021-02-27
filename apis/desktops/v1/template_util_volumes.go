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
	"strings"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// GetVolumes returns the volumes to mount to desktop pods.
func (t *Template) GetVolumes(cluster *appv1.VDICluster, desktop *Session, userdataVol string) []corev1.Volume {
	// Common volumes all containers will need.
	volumes := []corev1.Volume{
		{
			Name: v1.RunVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: v1.RunLockVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: v1.ShmVolume,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: v1.HostShmPath,
				},
			},
		},
		{
			Name: v1.TLSVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: desktop.GetName(),
				},
			},
		},
	}

	if t.NeedsEmptyTmpVolume() {
		volumes = append(volumes, corev1.Volume{
			Name: v1.TmpVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	if t.IsUNIXDisplaySocket() && !strings.HasPrefix(path.Dir(t.GetDisplaySocketAddress()), v1.DesktopTmpPath) {
		volumes = append(volumes, corev1.Volume{
			Name: v1.VNCSockVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	if t.NeedsDedicatedPulseVolume() {
		volumes = append(volumes, corev1.Volume{
			Name: v1.PulseSockVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// A PVC claim for the user if specified, otherwise use an EmptyDir.
	if userdataVol != "" {
		volumes = append(volumes, corev1.Volume{
			Name: v1.HomeVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: userdataVol,
				},
			},
		})
	} else {
		volumes = append(volumes, corev1.Volume{
			Name: v1.HomeVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// If systemd we need to add a few more temp filesystems and bind mount
	// /sys/fs/cgroup.
	if t.GetInitSystem() == InitSystemd || t.IsQEMUTemplate() {
		volumes = append(volumes, []corev1.Volume{
			{
				Name: v1.CgroupsVolume,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: v1.HostCgroupPath,
					},
				},
			},
		}...)
	}

	// Append the kvm device if a QEMU machine
	if t.IsQEMUTemplate() {
		volumes = append(volumes, corev1.Volume{
			Name: v1.KVMVolume,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: v1.DesktopKVMPath,
				},
			},
		})
		if t.QEMUUseCSI() {
			volumes = append(volumes, corev1.Volume{
				Name: v1.QEMUDiskVolume,
				VolumeSource: corev1.VolumeSource{
					CSI: &corev1.CSIVolumeSource{
						Driver: "image.csi.k8s.io",
						VolumeAttributes: map[string]string{
							"image":           t.GetQEMUDiskImage(),
							"imagePullPolicy": string(t.GetQEMUDiskImagePullPolicy()),
						},
					},
				},
			})
		}
	}

	if t.DindIsEnabled() {
		volumes = append(volumes, []corev1.Volume{
			{
				Name: v1.DockerDataVolume,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: v1.DockerBinVolume,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}...)
	}

	if len(t.Spec.Volumes) > 0 {
		volumes = append(volumes, t.Spec.Volumes...)
	}

	return volumes
}

// GetDesktopVolumeMounts returns the volume mounts for the main desktop container.
func (t *Template) GetDesktopVolumeMounts(cluster *appv1.VDICluster, desktop *Session) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      v1.RunVolume,
			MountPath: v1.DesktopRunPath,
		},
		{
			Name:      v1.RunLockVolume,
			MountPath: v1.DesktopRunLockPath,
		},
		{
			Name:      v1.ShmVolume,
			MountPath: v1.DesktopShmPath,
		},
		{
			Name:      v1.HomeVolume,
			MountPath: fmt.Sprintf(v1.DesktopHomeFmt, desktop.GetUser()),
		},
	}
	if t.NeedsEmptyTmpVolume() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.TmpVolume,
			MountPath: v1.DesktopTmpPath,
		})
	}
	if t.IsUNIXDisplaySocket() && !strings.HasPrefix(path.Dir(t.GetDisplaySocketAddress()), v1.DesktopTmpPath) {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.VNCSockVolume,
			MountPath: filepath.Dir(t.GetDisplaySocketAddress()),
		})
	}
	if t.NeedsDedicatedPulseVolume() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.PulseSockVolume,
			MountPath: filepath.Dir(t.GetPulseServer()),
		})
	}
	if t.GetInitSystem() == InitSystemd || t.IsQEMUTemplate() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.CgroupsVolume,
			MountPath: v1.DesktopCgroupPath,
		})
	}
	if t.IsQEMUTemplate() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.KVMVolume,
			MountPath: v1.DesktopKVMPath,
		})
		if t.QEMUUseCSI() {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      v1.QEMUDiskVolume,
				MountPath: v1.QEMUCSIDiskPath,
			})
		}
	}
	if t.DindIsEnabled() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.DockerBinVolume,
			MountPath: v1.DockerBinPath,
		})
	}
	if !t.IsQEMUTemplate() && t.Spec.DesktopConfig != nil && len(t.Spec.DesktopConfig.VolumeMounts) > 0 {
		mounts = append(mounts, t.Spec.DesktopConfig.VolumeMounts...)
	}
	return mounts
}

// NeedsEmptyTmpVolume returns true if none of the user-provided volumes provide
// the /tmp directory.
func (t *Template) NeedsEmptyTmpVolume() bool {
	if t.Spec.DesktopConfig == nil || len(t.Spec.DesktopConfig.VolumeMounts) == 0 {
		return true
	}
	for _, mount := range t.Spec.DesktopConfig.VolumeMounts {
		if strings.TrimSuffix(mount.MountPath, "/") == v1.DesktopTmpPath {
			return false
		}
	}
	return true
}

// GetTmpVolume returns the name of the volume providing the tmp directory.
func (t *Template) GetTmpVolume() string {
	if t.NeedsEmptyTmpVolume() {
		return v1.TmpVolume
	}
	for _, mount := range t.Spec.DesktopConfig.VolumeMounts {
		if strings.TrimSuffix(mount.MountPath, "/") == v1.DesktopTmpPath {
			return mount.Name
		}
	}
	return v1.TmpVolume
}
