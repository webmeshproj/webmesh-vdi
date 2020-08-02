package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/auth/mfa"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var apiLogger = logf.Log.WithName("api")

// DesktopAPI serves HTTP requests for the /api resource
type DesktopAPI interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// desktopAPI implements the DesktopAPI interface
type desktopAPI struct {
	// name of the vdi cluster
	clusterName string
	// the controller-runtime client
	client client.Client
	// the router interface
	router *mux.Router
	// our parent vdi cluster
	vdiCluster *v1alpha1.VDICluster
	// the user auth provider
	auth common.AuthProvider
	// the secrets backend
	secrets *secrets.SecretEngine
	// the mfa backend for setting and retrieving OTP secrets
	mfa *mfa.Manager
}

func (d *desktopAPI) handleClusterUpdate(req reconcile.Request) (reconcile.Result, error) {
	if req.NamespacedName.Name != d.clusterName {
		// ignore vdiclusters not tied to this app instance
		return reconcile.Result{}, nil
	}
	if d.vdiCluster == nil {
		// we are setting up the api the first time
		d.vdiCluster = &v1alpha1.VDICluster{}
		apiLogger.Info("Setting up kVDI runtime")
	} else {
		apiLogger.Info("Syncing kVDI runtime configuration with VDICluster spec")
	}

	var err error
	// overwrite the api vdicluster object with the remote state
	if err = d.client.Get(context.TODO(), req.NamespacedName, d.vdiCluster); err == nil {
		if d.secrets == nil {
			// we have not set up secrets yet
			d.secrets = secrets.GetSecretEngine(d.vdiCluster)
			// this means mfa also still need to be setup
			d.mfa = mfa.NewManager(d.secrets)
		}
		// call Setup on the secrets backend, should be idempotent
		if err := d.secrets.Setup(d.client, d.vdiCluster); err != nil {
			return reconcile.Result{}, err
		}
		if d.auth == nil {
			// auth has not been setup yet
			d.auth = auth.GetAuthProvider(d.vdiCluster, d.secrets)
		}
		// call Setup on the auth provider, should be idempotent
		if err := d.auth.Setup(d.client, d.vdiCluster); err != nil {
			return reconcile.Result{}, err
		}

	}
	return reconcile.Result{}, err
}

// NewFromConfig builds a new API router from the given kubernetes client configuration
// and vdi cluster name.
func NewFromConfig(cfg *rest.Config, vdiCluster string) (DesktopAPI, error) {
	// build our scheme
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	kclient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	// Create a manager for watching changes to the vdicluster configuration
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:         scheme,
		LeaderElection: false,
	})
	if err != nil {
		return nil, err
	}

	// create an api object
	api := &desktopAPI{
		clusterName: vdiCluster,
		client:      kclient,
	}

	// watch the vdicluster for updates, this also handles initial setup
	// of auth and secrets.
	var c controller.Controller
	if c, err = controller.New("cluster-watcher", mgr, controller.Options{
		Reconciler: reconcile.Func(api.handleClusterUpdate),
	}); err != nil {
		return nil, err
	}

	// set a watch on VDICluster objects
	if err = c.Watch(&source.Kind{Type: &v1alpha1.VDICluster{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}

	// start the mgr
	go func() {
		// we run this manager for life so no need to actually use this
		stop := make(chan struct{})
		// Start the manager. This will block until the stop channel is
		// closed, or the manager returns an error.
		if err := mgr.Start(stop); err != nil {
			apiLogger.Error(err, "VDICluster watcher died")
		}
	}()

	// return the api and build the router
	return api, api.buildRouter()
}
