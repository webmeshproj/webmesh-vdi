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
	"os"
	"path/filepath"
	"strings"

	jmespath "github.com/jmespath/go-jmespath"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
)

func checkClientInitErr(cmd *cobra.Command, args []string) error { return clientErr }

func notVersionCmd() bool {
	for _, arg := range os.Args {
		if arg == "version" {
			return false
		}
	}
	return true
}

func completeFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"json", "yaml"}, cobra.ShellCompDirectiveFilterFileExt
}

func completeVerbs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(rbacv1.VerbCreate),
		string(rbacv1.VerbRead),
		string(rbacv1.VerbUpdate),
		string(rbacv1.VerbDelete),
		string(rbacv1.VerbUse),
		string(rbacv1.VerbLaunch),
		string(rbacv1.VerbAll),
	}, cobra.ShellCompDirectiveDefault
}

func completeResources(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(rbacv1.ResourceUsers),
		string(rbacv1.ResourceRoles),
		string(rbacv1.ResourceTemplates),
		string(rbacv1.ResourceServiceAccounts),
		string(rbacv1.ResourceAll),
	}, cobra.ShellCompDirectiveDefault
}

func completeTemplates(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	templates, err := kvdiClient.GetDesktopTemplates()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}
	out := make([]string, 0)
	for _, tmpl := range templates {
		if !argsContains(args, tmpl.GetName()) {
			out = append(out, tmpl.GetName())
		}
	}
	return out, cobra.ShellCompDirectiveDefault
}

func completeRoles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	roles, err := kvdiClient.GetVDIRoles()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}
	out := make([]string, 0)
	for _, role := range roles {
		if !argsContains(args, role.GetName()) {
			out = append(out, role.GetName())
		}
	}
	return out, cobra.ShellCompDirectiveDefault
}

func completeSessions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	sessions, err := kvdiClient.GetDesktopSessions()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}
	out := make([]string, 0)
	for _, sess := range sessions.Sessions {
		out = append(out, sess.NamespacedName())
	}
	return out, cobra.ShellCompDirectiveDefault
}

func completeUsers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	users, err := kvdiClient.GetVDIUsers()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}
	out := make([]string, 0)
	for _, user := range users {
		if !argsContains(args, user.GetName()) {
			out = append(out, user.GetName())
		}
	}
	return out, cobra.ShellCompDirectiveDefault
}

func completeSessionPath(toComplete string) ([]string, cobra.ShellCompDirective) {
	spl := strings.Split(toComplete, ":")
	if len(spl) < 2 {
		return []string{}, cobra.ShellCompDirectiveError
	}
	sess := spl[0]
	nn, err := argToNamespacedName(sess)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}
	path := filepath.Dir(strings.Join(spl[1:], ":"))
	stat, err := kvdiClient.StatDesktopFile(nn, path)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}
	out := make([]string, 0)
	if stat.Stat.IsDirectory {
		for _, f := range stat.Stat.Contents {
			out = append(out, fmt.Sprintf("%s:%s/%s", sess, path, f.Name))
		}
	}
	return out, cobra.ShellCompDirectiveDefault
}

func writeObject(obj interface{}) error {
	var err error

	// Perform the search first if present
	if outFilter != "" {
		obj, err = jmespath.Search(outFilter, obj)
		if err != nil {
			return err
		}
	}

	// dump to json first to handle any json tags
	j, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	outFormat := viper.Get("server.output").(string)
	var out []byte

	switch outFormat {
	case "json":
		// return the raw json
		out = j
	case "yaml":
		// convert the json to yaml
		var in map[string]interface{}
		if err = json.Unmarshal(j, &in); err != nil {
			return err
		}
		out, err = yaml.Marshal(in)
	}

	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

func argsContains(args []string, s string) bool {
	for _, arg := range args {
		if arg == s {
			return true
		}
	}
	return false
}
