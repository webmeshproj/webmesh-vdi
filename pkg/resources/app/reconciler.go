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

package app

import (
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/pki"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler implements a reconciler for app-related resources.
type Reconciler struct {
	resources.VDIReconciler

	client client.Client
	scheme *runtime.Scheme
}

var _ resources.VDIReconciler = &Reconciler{}

// New returns a new App reconciler
func New(c client.Client, s *runtime.Scheme) *Reconciler {
	return &Reconciler{client: c, scheme: s}
}

// Reconcile reconciles all the core-components of a kVDI cluster.
func (f *Reconciler) Reconcile(reqLogger logr.Logger, instance *v1alpha1.VDICluster) error {
	// Generate the admin secret
	reqLogger.Info("Reconciling admin password secret")
	adminPass, err := f.reconcileAdminSecret(reqLogger, instance)
	if err != nil {
		return err
	}

	// Set up a temporary connection to the secrets engine
	reqLogger.Info("Setting up a temporary connection to the cluster secrets backend")
	secretsEngine := secrets.GetSecretEngine(instance)
	if err := secretsEngine.Setup(f.client, instance); err != nil {
		return err
	}
	defer func() {
		if err := secretsEngine.Close(); err != nil {
			reqLogger.Error(err, "Error cleaning up secrets engine")
		}
	}()

	// Reconcile a secret for generating JWT tokens
	reqLogger.Info("Reconciling JWT secrets")
	if _, err := secretsEngine.ReadSecret(v1.JWTSecretKey, false); err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return err
		}
		jwtSecret := common.GeneratePassword(32)
		if err := secretsEngine.WriteSecret(v1.JWTSecretKey, []byte(jwtSecret)); err != nil {
			return err
		}
	}

	reqLogger.Info("Reconciling built-in VDIRoles")
	// Reconcile the built-in roles.
	if err := reconcile.VDIRole(reqLogger, f.client, instance.GetAdminRole()); err != nil {
		return err
	}

	if err := reconcile.VDIRole(reqLogger, f.client, instance.GetLaunchTemplatesRole()); err != nil {
		return err
	}

	// reconcile any resources needed for the auth provider
	reqLogger.Info("Reconciling required resources for the configured authentication provider")
	authProvider := auth.GetAuthProvider(instance, secretsEngine)
	if err := authProvider.Reconcile(reqLogger, f.client, instance, adminPass); err != nil {
		return err
	}
	if err := authProvider.Close(); err != nil {
		reqLogger.Error(err, "Failed to close auth provider cleanly")
	}

	// Service account and cluster role/binding
	reqLogger.Info("Reconciling RBAC resources for the app servers")
	if err := reconcile.ServiceAccount(reqLogger, f.client, newAppServiceAccountForCR(instance)); err != nil {
		return err
	}
	if err := reconcile.ClusterRole(reqLogger, f.client, newAppClusterRoleForCR(instance)); err != nil {
		return err
	}
	if err := reconcile.ClusterRoleBinding(reqLogger, f.client, newRoleBindingsForCR(instance)); err != nil {
		return err
	}

	// Reconcile the PKI. This will ensure a CA as well as both server and client certificates
	// for the app deployment.
	reqLogger.Info("Reconciling PKI resources for mTLS")
	if err := pki.New(f.client, instance, secretsEngine).Reconcile(reqLogger); err != nil {
		return err
	}

	if instance.RunAppGrafanaSidecar() {
		// we need a configmap for grafana first
		reqLogger.Info("Reconciling grafana configuration")
		if err := reconcile.ConfigMap(reqLogger, f.client, newGrafanaConfigForCR(instance)); err != nil {
			return err
		}
	}

	// App deployment and service
	reqLogger.Info("Reconciling app deployment and services")
	if err := reconcile.Deployment(reqLogger, f.client, newAppDeploymentForCR(instance), true); err != nil {
		return err
	}
	if err := reconcile.Service(reqLogger, f.client, newAppServiceForCR(instance)); err != nil {
		return err
	}

	// Prometheus instance for aggregating metrics
	if instance.CreatePrometheusCR() {
		reqLogger.Info("Reconciling Prometheus deployment")
		if err := reconcile.Prometheus(reqLogger, f.client, newPrometheusForCR(instance)); err != nil {
			if ignoreNoPromOperator(reqLogger, err) != nil {
				return err
			}
		}
		if err := reconcile.Service(reqLogger, f.client, newPrometheusServiceForCR(instance)); err != nil {
			if ignoreNoPromOperator(reqLogger, err) != nil {
				return err
			}
		}
	}

	// ServiceMonitor for metrics scraping
	if instance.CreateAppServiceMonitor() {
		reqLogger.Info("Reconciling ServiceMonitor for app metrics")
		err = reconcile.ServiceMonitor(reqLogger, f.client, newAppServiceMonitorForCR(instance))
		return ignoreNoPromOperator(reqLogger, err)
	}

	return nil
}

func ignoreNoPromOperator(reqLogger logr.Logger, err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "no matches for kind") {
		reqLogger.Info("Could not create prometheus-operator object, is prometheus-operator installed?")
		return nil
	}
	return err
}
