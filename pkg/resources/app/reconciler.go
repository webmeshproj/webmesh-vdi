package app

import (
	"os"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AppClusterRole string

func init() {
	if role := os.Getenv("KVDI_APP_CLUSTER_ROLE"); role == "" {
		panic("No KVDI_APP_CLUSTER_ROLE set in the environment")
	} else {
		AppClusterRole = role
	}
}

type AppReconciler struct {
	resources.VDIReconciler

	client client.Client
	scheme *runtime.Scheme
}

var _ resources.VDIReconciler = &AppReconciler{}

// New returns a new App reconciler
func New(c client.Client, s *runtime.Scheme) resources.VDIReconciler {
	return &AppReconciler{client: c, scheme: s}
}

func (f *AppReconciler) Reconcile(reqLogger logr.Logger, instance *v1alpha1.VDICluster) error {
	// Service account and cluster role binding, role is created during deployment
	if err := reconcile.ReconcileServiceAccount(reqLogger, f.client, newAppServiceAccountForCR(instance)); err != nil {
		return err
	}
	if err := reconcile.ReconcileClusterRoleBinding(reqLogger, f.client, newRoleBindingsForCR(instance)); err != nil {
		return err
	}

	// Server certificate
	if err := reconcile.ReconcileCertificate(reqLogger, f.client, newAppCertForCR(instance), true); err != nil {
		return err
	}

	// Client certificate for novnc/db connections
	if err := reconcile.ReconcileCertificate(reqLogger, f.client, newAppClientCertForCR(instance), true); err != nil {
		return err
	}

	// App deployment and service
	if err := reconcile.ReconcileDeployment(reqLogger, f.client, newAppDeploymentForCR(instance), true); err != nil {
		return err
	}
	if err := reconcile.ReconcileService(reqLogger, f.client, newAppServiceForCR(instance)); err != nil {
		return err
	}
	return nil
}
