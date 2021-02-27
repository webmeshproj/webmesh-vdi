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
	"github.com/spf13/cobra"
)

func init() {
	sessionsCmd.AddCommand(sessionsGetCmd)

	rootCmd.AddCommand(sessionsCmd)
}

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"sess", "session", "s"},
	Short:   "Desktop sessions commands",
}

var sessionsGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Retrieve VDI sessions",
	Args:    cobra.NoArgs,
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := kvdiClient.GetDesktopSessions()
		if err != nil {
			return err
		}
		return writeObject(sessions.Sessions)
	},
}
