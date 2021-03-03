/*

   Copyright 2020,2021 Avi Zimmerman

   This file is part of kvdi.

   kvdi is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   kvdi is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with kvdi.  If not, see <https://www.gnu.org/licenses/>.

*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/tinyzimmer/kvdi/pkg/api/client"
)

// Internals
var (
	kvdiClient *client.Client
	clientErr  error
	cfgFile    string
	outFilter  string
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initClient)

	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.StringVarP(&cfgFile, "config", "c", "", `configuration file (default "$HOME/.kvdi.yaml")`)
	persistentFlags.StringP("server", "s", "https://127.0.0.1", "the address to the kvdi API server")
	persistentFlags.StringP("user", "u", "admin", "the username to use when authenticating against the API")
	persistentFlags.StringP("ca-file", "C", "", "the CA certificate to use to verify the API certificate")
	persistentFlags.BoolP("insecure-skip-verify", "k", false, "skip verification of the API server certificate")
	persistentFlags.StringP("output", "o", "json", "the format to dump results in")
	persistentFlags.StringVarP(&outFilter, "filter", "f", "", "a jmespath expression for filtering results (where applicable)")

	rootCmd.RegisterFlagCompletionFunc("output", completeFormats)
	rootCmd.MarkFlagFilename("config", "yaml", "yml", "json", "toml", "ini", "hcl", "env")
	rootCmd.MarkFlagFilename("ca-file", "crt", "pem")

	viper.BindPFlag("server.url", persistentFlags.Lookup("server"))
	viper.BindPFlag("server.user", persistentFlags.Lookup("user"))
	viper.BindPFlag("server.caFile", persistentFlags.Lookup("ca-file"))
	viper.BindPFlag("server.insecureSkipVerify", persistentFlags.Lookup("insecure-skip-verify"))
	viper.BindPFlag("server.output", persistentFlags.Lookup("output"))

	// Allow the configuration file to contain the actual certificate contents (base64 encoded)
	viper.SetDefault("server.caCert", "")
	// Allow the configuration file to contain a password
	viper.SetDefault("server.password", "")
}

// Execute executes the cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "kvdictl",
	Long: `kvdictl is a command line utility for interacting with the kvdi API server.

Configurations can be passed either on the command-line or via a viper configuration file.
The path to the configuration file defaults to ".kvdi.{json,toml,yaml,hcl,env}" in either the
current user's HOME directory or the current working directory (if HOME cannot be determined).
It can also be set explicitly using the -c or --config flags. Options passed on the CLI will 
override any of those provided in a configuration file.

There are two additional fields available in the configuration that are not exposed as flags.
Instead of a file, you can inline the CA certificate of the server directly with "server.caCert".
You may also specify the password for authentication at "server.password". If not found in the 
configuration file, you will be prompted when credentials are required. You may also set the 
password in the environment variable KVDI_PASSWORD to avoid being prompted. In the future, there 
will potentially be the ability to create API tokens specifically for API and CLI usage.

Using the CLI with a user that requires MFA is currently not supported.

An example for a configuration file might look similar to this:
   
    server:
      url: https://127.0.0.1
      user: admin
      password: "supersecret"
      insecureSkipVerify: false
      caFile: "/path/to/file.crt"
      # OR #
      caCert: |
        -----BEGIN CERTIFICATE-----
        MII...
        -----END CERTIFICATE-----

Most commands that provide an output can be configured to do so either in yaml or json. 
Additionally, the --filter flag can be used with a JMESpath to filter the output further.

Complete documentation for kvdi is available at https://github.com/kvdi/kvdi`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch viper.Get("server.output").(string) {
		case "json", "yaml":
			return nil
		default:
			return fmt.Errorf("%q is not a valid output format", viper.Get("server.output"))
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if clientErr == nil && kvdiClient != nil {
			kvdiClient.Close()
		}
	},
	SilenceUsage: true,
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		usr, err := user.Current()
		if err == nil {
			viper.AddConfigPath(usr.HomeDir)
		} else {
			path, err := os.Getwd()
			cobra.CheckErr(err)
			viper.AddConfigPath(path)
		}
		viper.SetConfigName(".kvdi")
	}

	viper.SetEnvPrefix("KVDI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return
		}
		// Any read related errors of a found configuration would
		// be fatal.
		fmt.Fprint(os.Stderr, "ERROR:", err.Error())
		os.Exit(2)
	}
}

func initClient() {
	var err error
	var tlsCA []byte
	var password []byte

	if caCertBody := viper.GetString("server.caCert"); caCertBody != "" {
		tlsCA = []byte(caCertBody)
	} else if caCertFile := viper.GetString("server.caFile"); caCertFile != "" {
		tlsCA, err = ioutil.ReadFile(caCertFile)
		cobra.CheckErr(err)
	}

	kvdiUser := viper.GetString("server.user")
	kvdiPassword := viper.GetString("server.password")

	if kvdiPassword == "" {
		kvdiPassword = os.Getenv("KVDI_PASSWORD")
	}

	if kvdiPassword == "" && notVersionCmd() {
		fmt.Printf("Enter Password for %q: ", kvdiUser)
		password, err = term.ReadPassword(int(os.Stdin.Fd()))
		cobra.CheckErr(err)
		kvdiPassword = string(password)
	}

	kvdiClient, clientErr = client.New(&client.Opts{
		URL:                   viper.GetString("server.url"),
		Username:              kvdiUser,
		Password:              kvdiPassword,
		TLSCACert:             tlsCA,
		TLSInsecureSkipVerify: viper.GetBool("server.insecureSkipVerify"),
	})

	// This would only happen during a bizarre memory allocation issue during cookiejar.New().
	// Authentication errors are not always fatal depending on the command being used, and the
	// client object will still be usable (e.g. when querying server version).
	if kvdiClient == nil {
		fmt.Fprint(os.Stderr, "ERROR: Fatal error creating kvdi client")
		os.Exit(3)
	}

	kvdiClient.SetAutoRefreshToken(false)
}
