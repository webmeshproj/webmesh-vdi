package app

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	// Generate the admin secret
	adminPass, err := f.reconcileAdminSecret(reqLogger, instance)
	if err != nil {
		return err
	}

	secretsProvider := secrets.GetProvider(instance)
	if err := secretsProvider.Setup(f.client, instance); err != nil {
		return err
	}

	if _, err := secretsProvider.ReadSecret(v1alpha1.JWTSecretKey, false); err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return err
		}
		jwtSecret := common.GeneratePassword(32)
		if err := secretsProvider.WriteSecret(v1alpha1.JWTSecretKey, []byte(jwtSecret)); err != nil {
			return err
		}
	}

	if err := reconcile.ReconcileVDIRole(reqLogger, f.client, instance.GetAdminRole()); err != nil {
		return err
	}

	if err := reconcile.ReconcileVDIRole(reqLogger, f.client, instance.GetLaunchTemplatesRole()); err != nil {
		return err
	}

	// reconcile any resources needed for the auth provider
	if err := auth.GetAuthProvider(instance).Reconcile(reqLogger, f.client, instance, adminPass); err != nil {
		return err
	}

	// Service account and cluster role/binding
	if err := reconcile.ReconcileServiceAccount(reqLogger, f.client, newAppServiceAccountForCR(instance)); err != nil {
		return err
	}
	if err := reconcile.ReconcileClusterRole(reqLogger, f.client, newAppClusterRoleForCR(instance)); err != nil {
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

	// The built-in admin role
	if err := reconcile.ReconcileVDIRole(reqLogger, f.client, instance.GetAdminRole()); err != nil {
		return err
	}

	return nil
}
