package api

import (
	"net/http"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var apiLogger = logf.Log.WithName("api")

// DesktopAPI serves HTTP requests for the /api resource
type DesktopAPI interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// desktopAPI implements the DesktopAPI interface
type desktopAPI struct {
	client     client.Client
	router     *mux.Router
	vdiCluster *v1alpha1.VDICluster
}

// NewFromConfig builds a new API router from the given kubernetes client configuration
// and vdi cluster name.
func NewFromConfig(cfg *rest.Config, vdiCluster string) (DesktopAPI, error) {
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return nil, err
	}
	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	apiLogger.Info("Retrieving VDICluster configuration")
	var found *v1alpha1.VDICluster
	for found == nil {
		if found, err = util.LookupClusterByName(client, vdiCluster); err != nil {
			apiLogger.Error(err, "Failed to retrieve VDICluster configuration, retrying in 2 seconds...")
			found = nil
			time.Sleep(time.Duration(2) * time.Second)
		}
	}
	api := &desktopAPI{
		client:     client,
		vdiCluster: found,
	}

	return api, api.buildRouter()
}
