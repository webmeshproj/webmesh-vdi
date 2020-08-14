package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/auth/mfa"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	util "github.com/tinyzimmer/kvdi/pkg/util/common"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func (d *desktopAPI) handleClusterUpdate(req reconcile.Request) error {
	if req.NamespacedName.Name != d.clusterName {
		// ignore vdiclusters not tied to this app instance
		return nil
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
	if err = d.client.Get(context.TODO(), req.NamespacedName, d.vdiCluster); err != nil {
		return err
	}

	if d.secrets == nil {
		// we have not set up secrets yet
		d.secrets = secrets.GetSecretEngine(d.vdiCluster)
		// this means mfa also still need to be setup
		d.mfa = mfa.NewManager(d.secrets)
	}
	// call Setup on the secrets backend, should be idempotent
	if err = d.secrets.Setup(d.client, d.vdiCluster); err != nil {
		return err
	}

	if d.auth == nil {
		// auth has not been setup yet
		d.auth = auth.GetAuthProvider(d.vdiCluster, d.secrets)
	}
	// call Setup on the auth provider, should be idempotent
	if err = d.auth.Setup(d.client, d.vdiCluster); err != nil {
		return err
	}

	return nil
}

func buildScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func getClientFromConfigAndScheme(cfg *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
	return client.New(cfg, client.Options{Scheme: scheme})
}

// NewFromConfig builds a new API router from the given kubernetes client configuration
// and vdi cluster name.
func NewFromConfig(cfg *rest.Config, vdiCluster string) (DesktopAPI, error) {
	// create an api object
	api := &desktopAPI{clusterName: vdiCluster}

	// build our scheme
	scheme, err := buildScheme()
	if err != nil {
		return nil, err
	}

	// build a client for routes to use
	api.client, err = getClientFromConfigAndScheme(cfg, scheme)
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

	// watch the vdicluster for updates, this also handles initial setup
	// of auth and secrets.
	var c controller.Controller
	if c, err = controller.New("cluster-watcher", mgr, controller.Options{
		Reconciler: reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
			return reconcile.Result{}, util.Retry(5, time.Second*2, func() error { return api.handleClusterUpdate(req) })
		}),
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

// NewTestAPI returns a new API using a fake kubernetes client and in-memory storage.
func NewTestAPI() (srvr *http.Server, addr, adminPass string, err error) {
	adminPass = "testing"

	// create an api object
	api := &desktopAPI{clusterName: "test-cluster"}

	// build our scheme
	var scheme *runtime.Scheme
	scheme, err = buildScheme()
	if err != nil {
		return
	}

	// build a client for routes to use
	api.client = fake.NewFakeClientWithScheme(scheme)

	// create a cluster object
	api.vdiCluster = &v1alpha1.VDICluster{}
	api.vdiCluster.Name = "test-cluster"
	if err = api.client.Create(context.TODO(), api.vdiCluster); err != nil {
		return
	}

	// create an admin role mapping
	if err = api.client.Create(context.TODO(), api.vdiCluster.GetAdminRole()); err != nil {
		return
	}

	// create a launch templates role
	if err = api.client.Create(context.TODO(), api.vdiCluster.GetLaunchTemplatesRole()); err != nil {
		return
	}

	// create a fake running pod/ns and set environment
	os.Setenv("POD_NAME", "test-server")
	os.Setenv("POD_NAMESPACE", "default")
	pod := &corev1.Pod{}
	pod.Name = "test-server"
	pod.Namespace = "default"
	if err = api.client.Create(context.TODO(), pod); err != nil {
		return
	}
	ns := &corev1.Namespace{}
	ns.Name = "default"
	if err = api.client.Create(context.TODO(), ns); err != nil {
		return
	}

	// build the api router
	if err = api.buildRouter(); err != nil {
		return
	}

	// set up auth and secrets
	api.secrets = secrets.GetSecretEngine(api.vdiCluster)
	api.mfa = mfa.NewManager(api.secrets)
	api.auth = auth.GetAuthProvider(api.vdiCluster, api.secrets)
	if err = api.secrets.Setup(api.client, api.vdiCluster); err != nil {
		return
	}
	if err = api.auth.Setup(api.client, api.vdiCluster); err != nil {
		return
	}

	// set a dummy jwt key
	if err = api.secrets.WriteSecret(v1.JWTSecretKey, []byte("supersecret")); err != nil {
		return
	}

	// reconcile initial credentials for auth
	// will be admin:testing
	if err = api.auth.Reconcile(apiLogger, api.client, api.vdiCluster, adminPass); err != nil {
		return
	}

	// build the base router
	r := mux.NewRouter()

	// add the api routes
	r.PathPrefix("/api").Handler(api)

	srvr = &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":0"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	netaddr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return
	}

	l, err := net.ListenTCP("tcp", netaddr)
	if err != nil {
		return
	}

	addr = fmt.Sprintf("http://127.0.0.1:%d", l.Addr().(*net.TCPAddr).Port)

	go func() {
		if err := srvr.Serve(l); err != nil {
			if err != http.ErrServerClosed {
				apiLogger.Error(err, "Error starting test server on local socket")
			}
		}
	}()

	return
}
