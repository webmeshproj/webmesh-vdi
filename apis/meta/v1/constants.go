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
	"path"
	"time"
)

const (
	// RoleClusterRefLabel marks for which cluster a role belongs
	RoleClusterRefLabel = "kvdi.io/cluster-ref"
	// CreationSpecAnnotation contains the serialized creation spec of a resource
	// to be compared against desired state.
	CreationSpecAnnotation = "kvdi.io/creation-spec"
	// LDAPGroupRoleAnnotation is an annotation applied to VDIRoles to "bind" them
	// to LDAP groups. A semicolon-separated list can bind a VDIRole to multiple
	// LDAP groups.
	LDAPGroupRoleAnnotation = "kvdi.io/ldap-groups"
	// OIDCGroupRoleAnnotation is the annotation applied to VDIRoles to "bind" them
	// to groups provided in claims from an OIDC provider. A semicolon separated list can
	// bind a role to multiple groups.
	OIDCGroupRoleAnnotation = "kvdi.io/oidc-groups"
	// AuthGroupSeparator is the separator used when parsing lists of groups from a string.
	AuthGroupSeparator = ";"
	// VDIClusterLabel is the label attached to resources to reference their parents VDI cluster
	VDIClusterLabel = "vdiCluster"
	// ComponentLabel is the label primarily used for service selectors
	ComponentLabel = "vdiComponent"
	// UserLabel is a label to tie the user id associated with a desktop instance
	UserLabel = "desktopUser"
	// DesktopNameLabel is a label referencing the name of the desktop instance. This is to add randomness
	// for the headless service selector placed in front of each pod.
	DesktopNameLabel = "desktopName"
	// ClientAddrLabel is the a label referencing the client address on a display/audio lock.
	ClientAddrLabel = "clientAddr"
	// ServerCertificateMountPath is where server certificates get placed inside pods
	ServerCertificateMountPath = "/etc/kvdi/tls/server"
	// ClientCertificateMountPath is where client certificates get placed inside pods
	ClientCertificateMountPath = "/etc/kvdi/tls/client"
	// SecretAssetsMountPath is a mount path for assets backed by secrets
	SecretAssetsMountPath = "/etc/kvdi/secrets"
	// JWTSecretKey is where our JWT secret is stored in the secrets backend.
	JWTSecretKey = "jwtSecret"
	// OTPUsersSecretKey is where a mapping of users to their OTP secrets is held in the secrets backend.
	OTPUsersSecretKey = "otpUsers"
	// RefreshTokensSecretKey is where a mapping of refresh tokens to users is kept in the secrets backend.
	RefreshTokensSecretKey = "refreshTokens"
	// WebPort is the port that web services will listen on internally
	WebPort = 8443
	// PublicWebPort is the port for the app service
	PublicWebPort = 443
	// DesktopRunDir is the dir mounted for internal runtime files
	DesktopRunDir = "/var/run/kvdi"
	// DefaultDisplaySocketAddr is the default path used for the display unix socket
	DefaultDisplaySocketAddr = "unix:///var/run/kvdi/display.sock"
	// DefaultNamespace is the default namespace to provision resources in
	DefaultNamespace = "default"
	// DefaultSessionLength is the session length used for setting expiry
	// times on new user sessions.
	DefaultSessionLength = time.Duration(15) * time.Minute
	// CACertKey is the key where the CA certificate is placed in TLS secrets.
	CACertKey = "ca.crt"
	// UserEnvVar is the environment variable used to set the username during a desktop's init
	// process
	UserEnvVar = "USER"
	// EnableRootEnvVar is the environment variable used to signal to the init process that
	// sudo access should be granted.
	EnableRootEnvVar = "ENABLE_ROOT"
	// VNCSockEnvVar is the environment variable used to set the VNC socket during the init
	// process.
	VNCSockEnvVar = "VNC_SOCK_ADDR"
	// UIDEnvVar is the environment varible where the UID of the user is set. This is a generic
	// UID used for all users.
	UIDEnvVar = "UID"
	// HomeEnvVar is the environment variable where the home directory of the user is set.
	HomeEnvVar = "HOME"
	// QEMUBootImageEnvVar contains the path to the root disk image for the virtual machine.
	QEMUBootImageEnvVar = "BOOT_IMAGE"
	// QEMUCloudImageEnvVar contains the path to the cloud-init image to use when booting the machine.
	QEMUCloudImageEnvVar = "CLOUD_IMAGE"
	// QEMUCPUsEnvVar contains the number of CPUs to allocate a virtual machine.
	QEMUCPUsEnvVar = "CPUS"
	// QEMUMemoryEnvVar contains the memory to allocate a virtual machine.
	QEMUMemoryEnvVar = "MEMORY"
)

// Desktop runtime volume names
var (
	TmpVolume        = "tmp"
	RunVolume        = "run"
	ShmVolume        = "shm"
	TLSVolume        = "tls"
	HomeVolume       = "home"
	CgroupsVolume    = "cgroups"
	RunLockVolume    = "run-lock"
	VNCSockVolume    = "vnc-sock"
	PulseSockVolume  = "pulse-sock"
	DockerDataVolume = "docker-data"
	DockerBinVolume  = "docker-bin"
	KVMVolume        = "qemu-kvm"
	QEMUDiskVolume   = "qemu-disk-image"
)

// Desktop runtime mount paths
const (
	HostShmPath    = "/dev/shm"
	HostCgroupPath = "/sys/fs/cgroup"

	DesktopTmpPath     = "/tmp"
	DesktopRunPath     = "/run"
	DesktopRunLockPath = "/run/lock"
	DesktopShmPath     = "/dev/shm"
	DesktopCgroupPath  = "/sys/fs/cgroup"
	DesktopHomeFmt     = "/home/%s"
	DesktopHomeMntPath = "/mnt/home"
	DesktopKVMPath     = "/dev/kvm"
	DockerDataPath     = "/var/lib/docker"
	DockerBinPath      = "/usr/local/docker/bin"
)

// Qemu variables
var (
	QEMUCSIDiskPath          = "/disk"
	QEMUNonCSIBootImagePath  = path.Join(DesktopRunPath, "boot.img")
	QEMUNonCSICloudImagePath = path.Join(DesktopRunPath, "cloud.img")
)

// Other defaults that we need to take the address of occasionally
var (
	DefaultUser     int64 = 9000
	DefaultReplicas int32 = 1
	True                  = true
	False                 = false
)
