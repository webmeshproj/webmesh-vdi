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

package desktops

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"

	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	"github.com/kvdi/kvdi/pkg/resources"
	"github.com/kvdi/kvdi/pkg/resources/desktop"
	"github.com/kvdi/kvdi/pkg/util/errors"
)

// SessionReconciler reconciles a Session object
type SessionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=pods;secrets;services;persistentvolumeclaims;persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=desktops.kvdi.io,resources=sessions;templates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=desktops.kvdi.io,resources=sessions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=desktops.kvdi.io,resources=sessions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SessionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("session", req.NamespacedName)

	reqLogger.Info("Reconciling Desktop")

	// Fetch the Desktop instance
	instance := &desktopsv1.Session{}
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

	reconcilers := []resources.DesktopReconciler{
		desktop.New(r.Client, r.Scheme),
	}

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
func (r *SessionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&desktopsv1.Session{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
