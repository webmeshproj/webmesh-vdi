package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/api"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
)

type LogOutput struct {
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"statusCode"`
	Size       int       `json:"size"`
	RemoteHost string    `json:"remoteHost"`
}

func formatLog(writer io.Writer, params handlers.LogFormatterParams) {
	host, _, err := net.SplitHostPort(params.Request.RemoteAddr)
	if err != nil {
		host = params.Request.RemoteAddr
	}
	if out, err := json.Marshal(&LogOutput{
		Time:       params.TimeStamp,
		Method:     params.Request.Method,
		Path:       params.URL.Path,
		StatusCode: params.StatusCode,
		RemoteHost: host,
		Size:       params.Size,
	}); err == nil {
		writer.Write(append(out, []byte("\n")...))
	}
}

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

	wrappedRouter := handlers.CompressHandler(handlers.CustomLoggingHandler(os.Stdout, r, formatLog))

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
