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
