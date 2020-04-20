package vdicluster

import (
	"context"
	"fmt"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/resources/app"
	"github.com/tinyzimmer/kvdi/pkg/resources/pki"
	"github.com/tinyzimmer/kvdi/pkg/resources/rethinkdb"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_vdicluster")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new VDICluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVDICluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("vdicluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VDICluster
	err = c.Watch(&source.Kind{Type: &v1alpha1.VDICluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployments and requeue the owner VDICluster
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.VDICluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSets and requeue the owner VDICluster
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.VDICluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Certificates and requeue the owner VDICluster
	err = c.Watch(&source.Kind{Type: &cm.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.VDICluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Secrets and requeue the owner VDICluster
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.VDICluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Services and requeue the owner VDICluster
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.VDICluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVDICluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVDICluster{}

// ReconcileVDICluster reconciles a VDICluster object
type ReconcileVDICluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a VDICluster object and makes changes based on the state read
// and what is in the VDICluster.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVDICluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VDICluster")

	// Fetch the VDICluster instance
	instance := &v1alpha1.VDICluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Build our reconcilers for this instance
	reconcilers := []resources.VDIReconciler{
		pki.New(r.client, r.scheme),
		rethinkdb.New(r.client, r.scheme),
		app.New(r.client, r.scheme),
	}

	// Run each reconciler
	for _, r := range reconcilers {
		if err := r.Reconcile(reqLogger, instance); err != nil {
			if qerr, ok := errors.IsRequeueError(err); ok {
				reqLogger.Info(fmt.Sprintf("Requeueing in %d seconds for: %s", qerr.Duration()/time.Second, qerr.Error()))
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: qerr.Duration(),
				}, nil
			}
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconcile finished")
	return reconcile.Result{}, nil
}
