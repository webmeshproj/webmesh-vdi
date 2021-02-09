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
	"context"
	"fmt"
	"time"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	rbacv1 "github.com/tinyzimmer/kvdi/apis/rbac/v1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/resources/app"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	kappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	krbacv1 "k8s.io/api/rbac/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// VDIClusterReconciler reconciles a VDICluster object
type VDIClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints;pods/log;configmaps;serviceaccounts;secrets;services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.kvdi.io,resources=vdiroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.kvdi.io,resources=vdiclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.kvdi.io,resources=vdiclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.kvdi.io,resources=vdiclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses;servicemonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates;issuers;clusterissuers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VDIClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("vdicluster", req.NamespacedName)

	reqLogger.Info("Reconciling VDICluster")

	// Fetch the VDICluster instance
	instance := &appv1.VDICluster{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Build our reconcilers for this instance
	reconcilers := []resources.VDIReconciler{
		// pki.New(r.client, r.scheme),
		app.New(r.Client, r.Scheme),
	}

	// Run each reconciler
	for _, r := range reconcilers {
		if err := r.Reconcile(ctx, reqLogger, instance); err != nil {
			if qerr, ok := errors.IsRequeueError(err); ok {
				reqLogger.Info(fmt.Sprintf("Requeueing in %d seconds for: %s", qerr.Duration()/time.Second, qerr.Error()))
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: qerr.Duration(),
				}, nil
			}
			return ctrl.Result{}, err
		}
	}

	reqLogger.Info("Reconcile finished")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VDIClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.VDICluster{}).
		Owns(&rbacv1.VDIRole{}).
		Owns(&kappsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&krbacv1.ClusterRole{}).
		Owns(&krbacv1.ClusterRoleBinding{}).
		Complete(r)
}
