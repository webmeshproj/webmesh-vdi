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
	"time"

	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
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
func (c *VDICluster) GetDesktopLabels(desktop *desktopsv1.Session) map[string]string {
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
