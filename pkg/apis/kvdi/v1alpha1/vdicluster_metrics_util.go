package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// CreateAppServiceMonitor returns true if the cluster specifies to create a
// ServiceMonitor.
func (c *VDICluster) CreateAppServiceMonitor() bool {
	if c.Spec.Metrics != nil && c.Spec.Metrics.ServiceMonitor != nil {
		return c.Spec.Metrics.ServiceMonitor.Create
	}
	return false
}

// CreatePrometheusCR returns true if the cluster specifies to create a
// Prometheus CR.
func (c *VDICluster) CreatePrometheusCR() bool {
	if c.Spec.Metrics != nil && c.Spec.Metrics.Prometheus != nil {
		return c.Spec.Metrics.Prometheus.Create
	}
	return false
}

// RunAppGrafanaSidecar returns true if a Grafana sidecar should be run alongside
// the app containers for displaying metrics in the UI.
func (c *VDICluster) RunAppGrafanaSidecar() bool {
	if c.Spec.Metrics != nil && c.Spec.Metrics.Grafana != nil {
		return c.Spec.Metrics.Grafana.Enabled
	}
	return false
}

// GetServiceMonitorLabels returns the labels to apply to the ServiceMonitor
// object.
func (c *VDICluster) GetServiceMonitorLabels() map[string]string {
	labels := c.GetComponentLabels("metrics")
	if c.Spec.Metrics != nil && c.Spec.Metrics.ServiceMonitor != nil {
		if len(c.Spec.Metrics.ServiceMonitor.Labels) > 0 {
			for k, v := range c.Spec.Metrics.ServiceMonitor.Labels {
				labels[k] = v
			}
			return labels
		}
	}
	labels["release"] = "prometheus"
	return labels
}

// GetPrometheusName returns the name of the kVDI prometheus deployment for this VDICluster.
func (c *VDICluster) GetPrometheusName() string {
	return fmt.Sprintf("%s-prometheus", c.GetName())
}

// GetPrometheusResources returns the resource requirements to place on the
// Prometheus deployment.
func (c *VDICluster) GetPrometheusResources() corev1.ResourceRequirements {
	if c.Spec.Metrics != nil && c.Spec.Metrics.Prometheus != nil {
		return c.Spec.Metrics.Prometheus.Resources
	}
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"memory": resource.MustParse("400Mi"),
		},
	}
}
