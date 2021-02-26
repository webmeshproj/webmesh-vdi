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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	serverConfigCmd.AddCommand(serverGetConfigCmd)

	configCmd.AddCommand(serverConfigCmd)

	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "kVDI configuration commands",
}

var serverConfigCmd = &cobra.Command{
	Use:     "server",
	Short:   "kVDI server configuration commands",
	PreRunE: checkClientInitErr,
}

var serverGetConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve server configurations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := kvdiClient.GetServerConfig()
		if err != nil {
			return err
		}

		j, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		var out []byte
		switch viper.Get("server.output").(string) {
		case "json":
			out = j
		case "yaml":
			var in map[string]interface{}
			err = json.Unmarshal(j, &in)
			if err != nil {
				return err
			}
			out, err = yaml.Marshal(in)
		}
		if err != nil {
			return err
		}

		fmt.Println(string(out))
		return nil
	},
}
