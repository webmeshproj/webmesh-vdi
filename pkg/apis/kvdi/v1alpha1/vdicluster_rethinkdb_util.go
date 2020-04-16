package v1alpha1

import (
	"fmt"
	"path/filepath"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *VDICluster) GetRethinkDBName() string {
	return fmt.Sprintf("%s-rethinkdb", c.GetName())
}

func (c *VDICluster) GetRethinkDBClientName() string {
	return fmt.Sprintf("%s-mgr", c.GetRethinkDBName())
}

func (c *VDICluster) GetRethinkDBProxyName() string {
	return fmt.Sprintf("%s-proxy", c.GetRethinkDBName())
}

func (c *VDICluster) GetRethinkDBImage() string {
	if c.Spec.RethinkDB != nil && c.Spec.RethinkDB.Image != "" {
		return c.Spec.RethinkDB.Image
	}
	return defaultRethinkDBImage
}

func (c *VDICluster) GetRethinkDBResources() corev1.ResourceRequirements {
	if c.Spec.RethinkDB != nil {
		return c.Spec.RethinkDB.DBResources
	}
	return corev1.ResourceRequirements{}
}

func (c *VDICluster) GetRethinkDBProxyResources() corev1.ResourceRequirements {
	if c.Spec.RethinkDB != nil {
		return c.Spec.RethinkDB.DBResources
	}
	return corev1.ResourceRequirements{}
}

func (c *VDICluster) GetRethinkDBReplicas() *int32 {
	if c.Spec.RethinkDB != nil && c.Spec.RethinkDB.Replicas != 0 {
		return &c.Spec.RethinkDB.Replicas
	}
	return &defaultRethinkDBReplicas
}

func (c *VDICluster) GetRethinkDBShards() *int32 {
	if c.Spec.RethinkDB != nil && c.Spec.RethinkDB.Shards != 0 {
		return &c.Spec.RethinkDB.Shards
	}
	return &defaultRethinkDBShards
}

func (c *VDICluster) GetRethinkDBProxyReplicas() *int32 {
	if c.Spec.RethinkDB != nil && c.Spec.RethinkDB.ProxyReplicas != 0 {
		return &c.Spec.RethinkDB.ProxyReplicas
	}
	return &defaultRethinkDBReplicas
}

// RethinkDBEnvVars returns the environment variables to supply to the
// rethinkdb pods.
func (c *VDICluster) RethinkDBEnvVars() []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name:  "TLS_CA_CERT",
			Value: filepath.Join(ServerCertificateMountPath, cmmeta.TLSCAKey),
		},
		{
			Name:  "TLS_CERT",
			Value: filepath.Join(ServerCertificateMountPath, corev1.TLSCertKey),
		},
		{
			Name:  "TLS_KEY",
			Value: filepath.Join(ServerCertificateMountPath, corev1.TLSPrivateKeyKey),
		},
		{
			Name:  "SERVICE_NAME",
			Value: c.GetRethinkDBName(),
		},
		{
			Name:  "PROXY_NAME",
			Value: c.GetRethinkDBProxyName(),
		},
	}
	return envVars
}

// GetRethinkDBVolumeClaims returns the PVC templates to supply to the RethinkDB
// StatefulSet.
func (c *VDICluster) GetRethinkDBVolumeClaims() []corev1.PersistentVolumeClaim {
	if c.Spec.RethinkDB != nil && c.Spec.RethinkDB.PVCSpec != nil {
		return []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-rethinkdb-storage", c.GetName()),
				},
				Spec: *c.Spec.RethinkDB.PVCSpec,
			},
		}
	}
	return []corev1.PersistentVolumeClaim{}
}

func (c *VDICluster) GetRethinkDBVolumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: fmt.Sprintf("%s-rethinkdb-tls", c.GetName()),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: c.GetRethinkDBName(),
				},
			},
		},
	}
	if c.Spec.RethinkDB != nil && c.Spec.RethinkDB.PVCSpec != nil {
		return volumes
	}
	return append(volumes, corev1.Volume{
		Name:         fmt.Sprintf("%s-rethinkdb-storage", c.GetName()),
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
}

// GetRethinkDBVolumeMounts returns the volume mounts to apply to the pods in the
// RethinkDB StatefulSet.
func (c *VDICluster) GetRethinkDBVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      fmt.Sprintf("%s-rethinkdb-tls", c.GetName()),
			MountPath: ServerCertificateMountPath,
		},
		{
			Name:      fmt.Sprintf("%s-rethinkdb-storage", c.GetName()),
			MountPath: "/data/rethinkdb_data",
		},
	}
}
