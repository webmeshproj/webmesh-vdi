package main

import (
	"fmt"
	"os"

	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var applogger = logf.Log.WithName("app")

func main() {
	var vdiCluster string
	var enableCORS bool
	pflag.CommandLine.StringVar(&vdiCluster, "vdi-cluster", "", "The VDICluster this application is serving")
	pflag.CommandLine.BoolVar(&enableCORS, "enable-cors", false, "Add CORS headers to requests")
	common.ParseFlagsAndSetupLogging()

	common.PrintVersion(applogger)

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		applogger.Error(err, "Failed to load kubernetes configuration")
		os.Exit(1)
	}

	// build the server
	srvr, err := newServer(cfg, vdiCluster, enableCORS)
	if err != nil {
		applogger.Error(err, "Failed to build the server router")
		os.Exit(1)
	}

	// serve
	applogger.Info(fmt.Sprintf("Starting VDI cluster frontend on :%d", v1.WebPort))
	if err := srvr.ListenAndServeTLS(tlsutil.ServerKeypair()); err != nil {
		applogger.Error(err, "Failed to start https server")
		os.Exit(1)
	}
}
