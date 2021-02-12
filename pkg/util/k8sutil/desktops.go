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

package k8sutil

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetVDIClusterForDesktop retrieves the VDICluster for this Desktop instance
func GetVDIClusterForDesktop(c client.Client, d *desktopsv1.Session) (*appv1.VDICluster, error) {
	nn := types.NamespacedName{Name: d.Spec.VDICluster, Namespace: metav1.NamespaceAll}
	found := &appv1.VDICluster{}
	return found, c.Get(context.TODO(), nn, found)
}

// GetDesktopVolumesFromTemplate returns the volumes to mount to desktop pods.
func GetDesktopVolumesFromTemplate(t *desktopsv1.Template, cluster *appv1.VDICluster, desktop *desktopsv1.Session) []corev1.Volume {
	// Common volumes all containers will need.
	volumes := []corev1.Volume{
		{
			Name: v1.TmpVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
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
	if cluster.GetUserdataVolumeSpec() != nil {
		volumes = append(volumes, corev1.Volume{
			Name: v1.HomeVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: cluster.GetUserdataVolumeName(desktop.GetUser()),
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
	if t.GetInitSystem() == desktopsv1.InitSystemd {
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

	if t.DindIsEnabled() {
		volumes = append(volumes, corev1.Volume{
			Name: v1.DockerDataVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
		volumes = append(volumes, corev1.Volume{
			Name: v1.DockerBinVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	if additionalVolumes := t.GetVolumes(); additionalVolumes != nil {
		volumes = append(volumes, additionalVolumes...)
	}

	return volumes
}

// GetDesktopVolumeMountsFromTemplate returns the volume mounts for the main desktop container.
func GetDesktopVolumeMountsFromTemplate(t *desktopsv1.Template, cluster *appv1.VDICluster, desktop *desktopsv1.Session) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
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
			Name:      v1.ShmVolume,
			MountPath: v1.DesktopShmPath,
		},
		{
			Name:      v1.HomeVolume,
			MountPath: fmt.Sprintf(v1.DesktopHomeFmt, desktop.GetUser()),
		},
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
	if t.GetInitSystem() == desktopsv1.InitSystemd {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.CgroupsVolume,
			MountPath: v1.DesktopCgroupPath,
		})
	}
	if t.DindIsEnabled() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v1.DockerBinVolume,
			MountPath: v1.DockerBinPath,
		})
	}
	if additionalMounts := t.GetVolumeMounts(); additionalMounts != nil {
		mounts = append(mounts, additionalMounts...)
	}
	return mounts
}
