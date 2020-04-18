package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/api"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
)

func newServer(cfg *rest.Config, vdiCluster string, enableCORS bool) (*http.Server, error) {
	// build the api router with our kubeconfig
	apiRouter, err := api.NewFromConfig(cfg, vdiCluster)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	// api routes
	r.PathPrefix("/api").Handler(apiRouter)
	// vue frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("/static")))

	wrappedRouter := handlers.CompressHandler(handlers.LoggingHandler(os.Stdout, r))

	if enableCORS {
		wrappedRouter = handlers.CORS()(wrappedRouter)
	}

	return &http.Server{
		Handler:      wrappedRouter,
		Addr:         fmt.Sprintf(":%d", v1alpha1.WebPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, nil
}
