package desktop

import (
	"context"
	"fmt"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/resources/desktop"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
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

var log = logf.Log.WithName("controller_desktop")

// Add creates a new Desktop Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDesktop{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("desktop-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Desktop
	err = c.Watch(&source.Kind{Type: &v1alpha1.Desktop{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Desktop
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Desktop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Desktop{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Certificates and requeue the owner Desktop
	err = c.Watch(&source.Kind{Type: &cm.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Desktop{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileDesktop implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileDesktop{}

// ReconcileDesktop reconciles a Desktop object
type ReconcileDesktop struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Desktop object and makes changes based on the state read
// and what is in the Desktop.Spec
func (r *ReconcileDesktop) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Desktop")

	// Fetch the Desktop instance
	instance := &v1alpha1.Desktop{}
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

	reconcilers := []resources.DesktopReconciler{
		desktop.New(r.client, r.scheme),
	}

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
