package cmd

import (
	// embeds bundle manifest
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tinyzimmer/kvdi/pkg/version"
)

//go:embed bundle.yaml
var bundleManifest string

var (
	managerVersion   string
	installNamespace string
)

func init() {
	installFlags := installCmd.Flags()

	installFlags.StringVar(&installNamespace, "namespace", "kvdi-system", "the namespace to use for the manifests")
	installFlags.StringVar(&managerVersion, "manager-version", version.Version, "the version of the kvdi-manager to install")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Output a manifest for installing or upgrading kVDI base resources",
	Long: `The install command can be used for install kVDI into a Kubernetes cluster.
	
The command outputs a complete manifest containing all of the CRDs for kVDI, along with
roles, serviceaccounts, services, and deployments for the manager.

Example Usage:

    kvdictl install | kubectl apply --validate=false -f -

`,
	RunE: func(cmd *cobra.Command, args []string) error {

		bundleManifest = strings.Replace(bundleManifest, "kvdi-system", installNamespace, -1)
		bundleManifest = strings.Replace(bundleManifest, version.Version, managerVersion, -1)

		fmt.Println(bundleManifest)
		return nil
	},
}
