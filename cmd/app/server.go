package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/api"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
)

type CustomFS struct {
}

func (s *CustomFS) isVNCAsset(name string) bool {
	switch filepath.Dir(name) {
	case "app", "core", "vendor":
		return true
	}
	return false
}

func (s *CustomFS) Open(name string) (http.File, error) {
	applogger.Info(name)
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := "/static"
	clean := filepath.FromSlash(path.Clean("/" + name))
	applogger.Info(clean)
	fullName := filepath.Join(dir, clean)
	f, err := os.Open(fullName)
	if err != nil {
		return nil, mapDirOpenError(err, fullName)
	}
	return f, nil
}

func mapDirOpenError(originalErr error, name string) error {
	if os.IsNotExist(originalErr) || os.IsPermission(originalErr) {
		return originalErr
	}
	parts := strings.Split(name, string(filepath.Separator))
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		fi, err := os.Stat(strings.Join(parts[:i+1], string(filepath.Separator)))
		if err != nil {
			return originalErr
		}
		if !fi.IsDir() {
			return os.ErrNotExist
		}
	}
	return originalErr
}

func newServer(cfg *rest.Config, vdiCluster string, enableCORS bool) (*http.Server, error) {
	// build the api router with our kubeconfig
	apiRouter, err := api.NewFromConfig(cfg, vdiCluster)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	// mTLS websocket proxy
	r.Path("/websockify/{endpoint}").HandlerFunc(mtlsWebsockify)
	// api routes
	r.PathPrefix("/api").Handler(apiRouter)
	// vue frontend
	r.PathPrefix("/").Handler(http.FileServer(&CustomFS{}))

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
