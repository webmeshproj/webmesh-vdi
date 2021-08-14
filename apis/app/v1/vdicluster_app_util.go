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

	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetAppName returns the name of the kVDI app deployment for this VDICluster.
func (c *VDICluster) GetAppName() string {
	return fmt.Sprintf("%s-app", c.GetName())
}

// GetServiceAnnotations returns the annotations to apply to the kvdi app service.
func (c *VDICluster) GetServiceAnnotations() map[string]string {
	annotations := c.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if c.Spec.App != nil && c.Spec.App.ServiceAnnotations != nil {
		for k, v := range c.Spec.App.ServiceAnnotations {
			annotations[k] = v
		}
	}
	return annotations
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
	return fmt.Sprintf("ghcr.io/kvdi/app:%s", version.Version)
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
		RunAsNonRoot: &v1.True,
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

// GetAppClientTLSSecretName returns the name of the client TLS secret for the app.
func (c *VDICluster) GetAppClientTLSSecretName() string {
	return fmt.Sprintf("%s-client", c.GetAppName())
}

// GetAppServerTLSSecretName returns the name of the server TLS secret for the app.
func (c *VDICluster) GetAppServerTLSSecretName() string {
	if c.Spec.App != nil && c.Spec.App.TLS != nil {
		if c.Spec.App.TLS.ServerSecret != "" {
			return c.Spec.App.TLS.ServerSecret
		}
	}
	return fmt.Sprintf("%s-server", c.GetAppName())
}

// AppIsUsingExternalServerTLS returns true if the app server certificate is user-supplied.
func (c *VDICluster) AppIsUsingExternalServerTLS() bool {
	if c.Spec.App != nil && c.Spec.App.TLS != nil {
		return c.Spec.App.TLS.ServerSecret != ""
	}
	return false
}

// GetAppClientTLSNamespacedName returns the namespaced name for the client TLS certificate.
func (c *VDICluster) GetAppClientTLSNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      c.GetAppClientTLSSecretName(),
		Namespace: c.GetCoreNamespace(),
	}
}

// GetAppServerTLSNamespacedName returns the namespaced name for the server TLS certificate.
func (c *VDICluster) GetAppServerTLSNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      c.GetAppServerTLSSecretName(),
		Namespace: c.GetCoreNamespace(),
	}
}
