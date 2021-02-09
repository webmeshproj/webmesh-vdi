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

// Package resources contains the interfaces for resource reconcilers to implement.
package resources

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/go-logr/logr"
)

// VDIReconciler represents an interface for ensuring resources for a VDI cluster
type VDIReconciler interface {
	Reconcile(logr.Logger, *v1alpha1.VDICluster) error
}

// VDIClusterReconcileFunc is a function for reconciling vdi cluster resources
type VDIClusterReconcileFunc func(logr.Logger, *v1alpha1.VDICluster) error

// DesktopReconciler represents an interface for ensuring resources for a
// single desktop instance.
type DesktopReconciler interface {
	Reconcile(logr.Logger, *v1alpha1.Desktop) error
}

// DesktopClusterReconcileFunc is a function for reconciling desktop resources.
type DesktopClusterReconcileFunc func(logr.Logger, *v1alpha1.Desktop) error
