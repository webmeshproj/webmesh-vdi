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

	metav1 "github.com/kvdi/kvdi/apis/meta/v1"
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/spf13/cobra"
)

var (
	createRoleOpts       types.CreateRoleRequest
	updateRoleName       string
	ruleVerbs            []string
	ruleResources        []string
	ruleResourcePatterns []string
	ruleNamespaces       []string
)

func init() {
	createFlags := roleCreateCmd.Flags()
	createFlags.StringVar(&createRoleOpts.Name, "name", "", "the name to assign the new role")
	createFlags.StringToStringVar(&createRoleOpts.Annotations, "annotations", map[string]string{}, "annotations to apply to the role")
	roleCreateCmd.MarkFlagRequired("name")

	addUpdateRoleFlags(roleRulesAddCmd)
	addUpdateRoleFlags(roleRulesRemoveCmd)
	addUpdateRoleFlags(roleAnnotationsCmd)
	addRuleFlags(roleRulesAddCmd)
	addRuleFlags(roleRulesRemoveCmd)
	addRuleFlags(roleCreateCmd)

	roleRulesCmd.AddCommand(roleRulesAddCmd)
	roleRulesCmd.AddCommand(roleRulesRemoveCmd)

	roleAnnotationsCmd.AddCommand(roleAnnotationsSetCmd)
	roleAnnotationsCmd.AddCommand(roleAnnotationsRemoveCmd)

	rolesCmd.AddCommand(rolesGetCmd)
	rolesCmd.AddCommand(roleCreateCmd)
	rolesCmd.AddCommand(rolesDeleteCmd)
	rolesCmd.AddCommand(roleRulesCmd)
	rolesCmd.AddCommand(roleAnnotationsCmd)

	rootCmd.AddCommand(rolesCmd)
}

func addUpdateRoleFlags(cmd *cobra.Command) {
	flagSet := cmd.PersistentFlags()

	flagSet.StringVar(&updateRoleName, "name", "", "the role to apply the changes to")
	cmd.MarkFlagRequired("name")
	cmd.RegisterFlagCompletionFunc("name", completeRoles)
}

func addRuleFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringSliceVar(&ruleVerbs, "verbs", []string{}, "verbs for the rule")
	flagSet.StringSliceVar(&ruleResources, "resources", []string{}, "resources for a rule")
	flagSet.StringSliceVar(&ruleResourcePatterns, "resource-patterns", []string{}, "resource patterns for the rule")
	flagSet.StringSliceVar(&ruleNamespaces, "namespaces", []string{}, "namespaces for the rule")

	cmd.RegisterFlagCompletionFunc("verbs", completeVerbs)
	cmd.RegisterFlagCompletionFunc("resources", completeResources)
}

func ruleFlagsToRule() rbacv1.Rule {
	r := rbacv1.Rule{
		ResourcePatterns: ruleResourcePatterns,
		Namespaces:       ruleNamespaces,
	}
	if len(ruleVerbs) > 0 {
		verbs := make([]rbacv1.Verb, len(ruleVerbs))
		for i, verb := range ruleVerbs {
			verbs[i] = rbacv1.Verb(verb)
		}
		r.Verbs = verbs
	}
	if len(ruleResources) > 0 {
		resources := make([]rbacv1.Resource, len(ruleResources))
		for i, resource := range ruleResources {
			resources[i] = rbacv1.Resource(resource)
		}
		r.Resources = resources
	}
	return r
}

var rolesCmd = &cobra.Command{
	Use:     "roles",
	Aliases: []string{"role", "r"},
	Short:   "Roles commands",
}

var rolesGetCmd = &cobra.Command{
	Use:               "get [NAME]",
	Short:             "Retrieve VDI role(s)",
	Args:              cobra.MaximumNArgs(1),
	PreRunE:           checkClientInitErr,
	ValidArgsFunction: completeRoles,
	RunE: func(cmd *cobra.Command, args []string) error {
		var out interface{}
		var err error
		if len(args) == 1 {
			out, err = kvdiClient.GetVDIRole(args[0])
		} else {
			out, err = kvdiClient.GetVDIRoles()
		}
		if err != nil {
			return err
		}
		return writeObject(out)
	},
}

var roleCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create VDI roles",
	Args:    cobra.NoArgs,
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		createRoleOpts.Rules = []rbacv1.Rule{ruleFlagsToRule()}
		if err := kvdiClient.CreateVDIRole(&createRoleOpts); err != nil {
			return err
		}
		fmt.Printf("Role %q created successfully\n", createRoleOpts.Name)
		return nil
	},
}

var rolesDeleteCmd = &cobra.Command{
	Use:               "delete [ROLES...]",
	Aliases:           []string{"del", "remove", "rem", "rm"},
	Short:             "Delete VDI roles",
	ValidArgsFunction: completeRoles,
	PreRunE:           checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if err := kvdiClient.DeleteVDIRole(arg); err != nil {
				return err
			}
			fmt.Printf("Role %q deleted successfully\n", arg)
		}
		return nil
	},
}

var roleRulesCmd = &cobra.Command{
	Use:     "rules",
	Aliases: []string{"rule"},
	Short:   "Manage VDI role rules",
}

var roleAnnotationsCmd = &cobra.Command{
	Use:     "annotations",
	Aliases: []string{"annotate"},
	Short:   "Manage VDI role annotations",
}

var roleRulesAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "create"},
	Short:   "Add rules to a VDI role",
	Args:    cobra.NoArgs,
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateRoleName == "" {
			return errors.New("You must provide a role name")
		}
		newRule := ruleFlagsToRule()
		if newRule.IsEmpty() {
			return errors.New("You must specify fields for the rule")
		}
		role, err := kvdiClient.GetVDIRole(updateRoleName)
		if err != nil {
			return err
		}
		role.Rules = append(role.Rules, newRule)
		if err := kvdiClient.UpdateVDIRole(updateRoleName, &types.UpdateRoleRequest{
			Annotations: role.GetAnnotations(),
			Rules:       role.Rules,
		}); err != nil {
			return err
		}
		fmt.Printf("Role %q updated successfully\n", updateRoleName)
		return nil
	},
}

var roleRulesRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rem", "rm", "delete", "del"},
	Short:   "Remove rules from a VDI role",
	Long: `Removes a rule from a VDI role.

The rule definition passed on the command-line must match the rule for removal exactly.
If multiple matches are found, they will all be removed.`,
	Args:    cobra.NoArgs,
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateRoleName == "" {
			return errors.New("You must provide a role name")
		}
		role, err := kvdiClient.GetVDIRole(updateRoleName)
		if err != nil {
			return err
		}
		toDelete := ruleFlagsToRule()
		opts := &types.UpdateRoleRequest{
			Annotations: role.GetAnnotations(),
			Rules:       make([]rbacv1.Rule, 0),
		}
		var ruleRemoved bool
		for _, rule := range role.Rules {
			if !toDelete.DeepEqual(rule) {
				opts.Rules = append(opts.Rules, rule)
				continue
			}
			ruleRemoved = true
		}
		if !ruleRemoved {
			return fmt.Errorf("No rules for %q matched the one provided for deletion", updateRoleName)
		}
		if err := kvdiClient.UpdateVDIRole(updateRoleName, opts); err != nil {
			return err
		}
		fmt.Printf("Role %q updated successfully\n", updateRoleName)
		return nil
	},
}

var roleAnnotationsSetCmd = &cobra.Command{
	Use:     "set",
	Aliases: []string{"add"},
	Short:   "Set annotations on a VDI role",
	Args:    cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		avail := []string{metav1.LDAPGroupRoleAnnotation, metav1.OIDCGroupRoleAnnotation}
		role, err := kvdiClient.GetVDIRole(updateRoleName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}
		annotations := role.GetAnnotations()
		if annotations == nil {
			return avail, cobra.ShellCompDirectiveDefault
		}
		for key := range annotations {
			if !argsContains(avail, key) {
				avail = append(avail, key)
			}
		}
		return avail, cobra.ShellCompDirectiveDefault
	},
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateRoleName == "" {
			return errors.New("You must provide a role name")
		}
		role, err := kvdiClient.GetVDIRole(updateRoleName)
		if err != nil {
			return err
		}
		annotations := role.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[args[0]] = args[1]
		opts := &types.UpdateRoleRequest{
			Rules:       role.Rules,
			Annotations: annotations,
		}
		if err := kvdiClient.UpdateVDIRole(updateRoleName, opts); err != nil {
			return err
		}
		fmt.Printf("Role %q updated successfully\n", updateRoleName)
		return nil
	},
}

var roleAnnotationsRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rem", "rm", "delete", "del", "unset"},
	Short:   "Remove annotations on a VDI role",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		role, err := kvdiClient.GetVDIRole(updateRoleName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}
		annotations := role.GetAnnotations()
		if annotations == nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		keys := make([]string, 0)
		for key := range annotations {
			keys = append(keys, key)
		}
		return keys, cobra.ShellCompDirectiveDefault
	},
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateRoleName == "" {
			return errors.New("You must provide a role name")
		}
		role, err := kvdiClient.GetVDIRole(updateRoleName)
		if err != nil {
			return err
		}
		annotations := role.GetAnnotations()
		if annotations == nil {
			return fmt.Errorf("Role %q does not have any annotations", updateRoleName)
		}
		if _, ok := annotations[args[0]]; !ok {
			return fmt.Errorf("Role %q does not contain a %q annotation", updateRoleName, args[0])
		}
		delete(annotations, args[0])
		opts := &types.UpdateRoleRequest{
			Rules:       role.Rules,
			Annotations: annotations,
		}
		if err := kvdiClient.UpdateVDIRole(updateRoleName, opts); err != nil {
			return err
		}
		fmt.Printf("Role %q updated successfully\n", updateRoleName)
		return nil
	},
}
