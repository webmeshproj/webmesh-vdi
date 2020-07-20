package v1alpha1

import (
	"fmt"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/version"

	corev1 "k8s.io/api/core/v1"
)

// GetAppName returns the name of the kVDI app deployment for this VDICluster.
func (c *VDICluster) GetAppName() string {
	return fmt.Sprintf("%s-app", c.GetName())
}

// GetAppReplicas returns the number of app replicas to run in this VDICluster.
// TODO: auto-scaling?
func (c *VDICluster) GetAppReplicas() *int32 {
	if c.Spec.App != nil && c.Spec.App.Replicas != 0 {
		return &c.Spec.App.Replicas
	}
	return &v1.DefaultReplicas
}

// GetAppResources returns the resource requirements for the app deployments.
func (c *VDICluster) GetAppResources() corev1.ResourceRequirements {
	if c.Spec.App != nil {
		return c.Spec.App.Resources
	}
	return corev1.ResourceRequirements{}
}

// GetAppImage returns the image to use for the app deployment.
func (c *VDICluster) GetAppImage() string {
	if c.Spec.App != nil && c.Spec.App.Image != "" {
		return c.Spec.App.Image
	}
	return fmt.Sprintf("quay.io/tinyzimmer/kvdi:app-%s", version.Version)
}

// GetAppPullPolicy returns the ImagePullPolicy to use for the app deployment.
func (c *VDICluster) GetAppPullPolicy() corev1.PullPolicy {
	return corev1.PullIfNotPresent
}

// GetAppSecurityContext returns the pod security context for the app deployment.
func (c *VDICluster) GetAppSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    &v1.DefaultUser,
		RunAsGroup:   &v1.DefaultUser,
		RunAsNonRoot: &v1.TrueVal,
	}
}

// EnableCORS returns true if CORS headers should be included in responses from the
// app server.
func (c *VDICluster) EnableCORS() bool {
	if c.Spec.App != nil {
		return c.Spec.App.CORSEnabled
	}
	return false
}

// AuditLogEnabled returns true if auditing events should be logged to stdout.
func (c *VDICluster) AuditLogEnabled() bool {
	if c.Spec.App != nil {
		return c.Spec.App.AuditLog
	}
	return false
}

// GetAppSecretsName returns the name of the secret to use for app secrets.
func (c *VDICluster) GetAppSecretsName() string {
	if c.Spec.Secrets != nil && c.Spec.Secrets.K8SSecret != nil && c.Spec.Secrets.K8SSecret.SecretName != "" {
		return c.Spec.Secrets.K8SSecret.SecretName
	}
	return fmt.Sprintf("%s-app-secrets", c.GetName())
}
