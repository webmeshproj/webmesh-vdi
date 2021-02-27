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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

var (
	userCreateOpts     types.CreateUserRequest
	userUpdateOpts     types.UpdateUserRequest
	userUpdateGenPassw bool
	userPasswLength    int
)

func init() {
	createFlags := userCreateCmd.Flags()
	createFlags.StringVar(&userCreateOpts.Username, "name", "", "the name for the new user")
	createFlags.StringVar(&userCreateOpts.Password, "password", "", "the password for the user, one will be generated and printed to the console if unset")
	createFlags.StringSliceVar(&userCreateOpts.Roles, "roles", []string{}, "roles to assign the new user")
	createFlags.IntVar(&userPasswLength, "password-length", 16, "the length to use when generating passwords")
	userCreateCmd.MarkFlagRequired("name")
	userCreateCmd.MarkFlagRequired("roles")
	userCreateCmd.RegisterFlagCompletionFunc("roles", completeRoles)

	updateFlags := userUpdateCmd.Flags()
	updateFlags.StringVar(&userUpdateOpts.Password, "password", "", "update the password for the user")
	updateFlags.StringSliceVar(&userUpdateOpts.Roles, "roles", []string{}, "update the roles for the user")
	updateFlags.BoolVar(&userUpdateGenPassw, "generate-password", false, "generate a new password to replace the existing one with")
	updateFlags.IntVar(&userPasswLength, "password-length", 16, "the length to use when generating passwords")
	userUpdateCmd.RegisterFlagCompletionFunc("roles", completeRoles)

	usersCmd.AddCommand(usersGetCmd)
	usersCmd.AddCommand(userCreateCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	usersCmd.AddCommand(userUpdateCmd)

	rootCmd.AddCommand(usersCmd)
}

var usersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "usr", "u"},
	Short:   "Users commands",
}

var usersGetCmd = &cobra.Command{
	Use:               "get [<NAME>]",
	Short:             "Retrieve VDI user(s)",
	Args:              cobra.MaximumNArgs(1),
	PreRunE:           checkClientInitErr,
	ValidArgsFunction: completeUsers,
	RunE: func(cmd *cobra.Command, args []string) error {
		var out interface{}
		var err error
		if len(args) == 1 {
			out, err = kvdiClient.GetVDIUser(args[0])
		} else {
			out, err = kvdiClient.GetVDIUsers()
		}
		if err != nil {
			return err
		}
		return writeObject(out)
	},
}

var userCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create VDI users",
	Args:    cobra.NoArgs,
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		var generatedPass string
		if userCreateOpts.Password == "" {
			generatedPass = common.GeneratePassword(userPasswLength)
			userCreateOpts.Password = generatedPass
		}
		if err := kvdiClient.CreateVDIUser(&userCreateOpts); err != nil {
			return err
		}
		fmt.Printf("User %q created successfully\n", userCreateOpts.Username)
		if generatedPass != "" {
			fmt.Println("  Generated Password:", generatedPass)
		}
		return nil
	},
}

var usersDeleteCmd = &cobra.Command{
	Use:               "delete [USERS...]",
	Aliases:           []string{"del", "remove", "rem", "rm"},
	Short:             "Delete VDI users",
	ValidArgsFunction: completeUsers,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if clientErr != nil {
			return clientErr
		}
		for _, arg := range args {
			if arg == "admin" {
				return errors.New(`You cannot delete the "admin" user`)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if err := kvdiClient.DeleteVDIUser(arg); err != nil {
				return err
			}
			fmt.Printf("User %q deleted successfully\n", arg)
		}
		return nil
	},
}

var userUpdateCmd = &cobra.Command{
	Use:               "update [USER]",
	Short:             "Update VDI users",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeUsers,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if clientErr != nil {
			return clientErr
		}
		for _, arg := range args {
			if arg == "admin" {
				return errors.New(`You cannot update the "admin" user`)
			}
		}
		if userUpdateOpts.Password == "" && len(userUpdateOpts.Roles) == 0 && !userUpdateGenPassw {
			return errors.New("No changes specified")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		var generatedPass string
		if userUpdateOpts.Password == "" && len(userUpdateOpts.Roles) == 0 {
			generatedPass = common.GeneratePassword(userPasswLength)
			userUpdateOpts.Password = generatedPass
		}
		if err := kvdiClient.UpdateVDIUser(username, &userUpdateOpts); err != nil {
			return err
		}
		fmt.Printf("User %q updated successfully\n", args[0])
		if generatedPass != "" {
			fmt.Println("  Generated Password:", generatedPass)
		}
		return nil
	},
}
