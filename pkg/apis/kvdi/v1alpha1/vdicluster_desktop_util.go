package v1alpha1

import (
	"time"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

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

// GetMaxSessionsPerUser returns the maximum number of sessions a user can run for this VDICluster.
func (c *VDICluster) GetMaxSessionsPerUser() int {
	if c.Spec.Desktops != nil {
		return c.Spec.Desktops.SessionsPerUser
	}
	return 0
}

// GetUserDesktopSelector returns a selector that can be used to find desktops for a given user.
func (c *VDICluster) GetUserDesktopSelector(username string) map[string]string {
	return map[string]string{
		v1.UserLabel:       username,
		v1.VDIClusterLabel: c.GetName(),
	}
}

// GetDesktopLabels returns the labels to apply to components for a desktop.
func (c *VDICluster) GetDesktopLabels(desktop *Desktop) map[string]string {
	labels := desktop.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[v1.UserLabel] = desktop.GetUser()
	labels[v1.VDIClusterLabel] = c.GetName()
	labels[v1.ComponentLabel] = "desktop"
	labels[v1.DesktopNameLabel] = desktop.GetName()
	return labels
}
