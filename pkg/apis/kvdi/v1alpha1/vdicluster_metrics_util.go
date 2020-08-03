package v1alpha1

// CreateAppServiceMonitor returns true if the CRD specifies to create a
// ServiceMonitor.
func (c *VDICluster) CreateAppServiceMonitor() bool {
	if c.Spec.Metrics != nil && c.Spec.Metrics.ServiceMonitor != nil {
		return c.Spec.Metrics.ServiceMonitor.Create
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
