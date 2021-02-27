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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	clientConfigCmd.AddCommand(setClientConfigCmd)
	clientConfigCmd.AddCommand(getClientConfigCmd)

	configCmd.AddCommand(serverConfigCmd)
	configCmd.AddCommand(clientConfigCmd)

	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"conf", "c"},
	Short:   "Configuration commands",
}

var serverConfigCmd = &cobra.Command{
	Use:     "server",
	Short:   "Retrieve server configurations",
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := kvdiClient.GetServerConfig()
		if err != nil {
			return err
		}
		return writeObject(cfg)
	},
}

// TODO: Allow configuring server?

var clientConfigCmd = &cobra.Command{
	Use:   "client",
	Short: "Client configuration commands",
}

var getClientConfigCmd = &cobra.Command{
	Use:   "get <PATH>",
	Short: "Retrieve client configurations",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return viper.AllKeys(), cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) error { return writeObject(viper.Get(args[0])) },
}

var setClientConfigCmd = &cobra.Command{
	Use:   "set <PATH> <VALUE>",
	Short: "Set client configurations",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return viper.AllKeys(), cobra.ShellCompDirectiveDefault
		}
		return []string{}, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		switch strings.ToLower(value) {
		case "true":
			viper.Set(key, true)
		case "false":
			viper.Set(key, false)
		default:
			viper.Set(key, value)
		}
		return viper.WriteConfig()
	},
}
