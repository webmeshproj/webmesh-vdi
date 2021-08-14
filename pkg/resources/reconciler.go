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
	"context"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"

	"github.com/go-logr/logr"
)

// VDIReconciler represents an interface for ensuring resources for a VDI cluster
type VDIReconciler interface {
	Reconcile(context.Context, logr.Logger, *appv1.VDICluster) error
}

// DesktopReconciler represents an interface for ensuring resources for a
// single desktop instance.
type DesktopReconciler interface {
	Reconcile(context.Context, logr.Logger, *desktopsv1.Session) error
}
