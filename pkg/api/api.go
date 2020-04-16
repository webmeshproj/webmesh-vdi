package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DesktopAPI serves HTTP requests for the /api resource
type DesktopAPI interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// desktopAPI implements the DesktopAPI interface
type desktopAPI struct {
	client     client.Client
	router     *mux.Router
	vdiCluster string
}

// NewFromConfig builds a new API router from the given kubernetes client configuration
// and vdi cluster name.
// TODO: The manager init is hacky right now, and really we should be pulling
// the entire VDICluster object into memory.
func NewFromConfig(cfg *rest.Config, vdiCluster string) (DesktopAPI, error) {
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:      metav1.NamespaceAll,
		LeaderElection: false,
	})
	if err != nil {
		return nil, err
	}
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}
	go func() {
		if err := mgr.Start(nil); err != nil {
			panic(err)
		}
	}()
	api := &desktopAPI{
		client:     mgr.GetClient(),
		vdiCluster: vdiCluster,
	}
	api.buildRouter()
	return api, nil
}
