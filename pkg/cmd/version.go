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

	"github.com/spf13/cobra"
	"github.com/tinyzimmer/kvdi/pkg/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieve kVDI version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Client Version:", version.Version)
		fmt.Println("    Git Commit:", version.GitCommit)

		serverVers, serverCommit, err := kvdiClient.GetServerVersion()
		if err != nil {
			fmt.Println("Error retrieving server information:", err.Error())
			return
		}

		fmt.Println("Server Version:", serverVers)
		fmt.Println("    Git Commit:", serverCommit)
	},
}
