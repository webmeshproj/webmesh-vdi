package v1alpha1

import "time"

// GetMaxSessionLength returns the duration to wait to kill a desktop pod.
// If the duration is not parseable or unconfigured, 0 is returned.
func (c *VDICluster) GetMaxSessionLength() time.Duration {
	if c.Spec.Desktops != nil && c.Spec.Desktops.MaxSessionLength != "" {
		dur, err := time.ParseDuration(c.Spec.Desktops.MaxSessionLength)
		if err != nil {
			return time.Duration(0)
		}
		return dur
	}
	return time.Duration(0)
}
