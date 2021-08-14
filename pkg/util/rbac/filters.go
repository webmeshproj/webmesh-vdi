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

package rbac

import (
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/types"
)

// FilterTemplates will take a list of DesktopTemplates and filter them based
// off which ones the user is allowed to use.
func FilterTemplates(u *types.VDIUser, tmpls []*desktopsv1.Template) []*desktopsv1.Template {
	filtered := make([]*desktopsv1.Template, 0)
	for _, tmpl := range tmpls {
		action := &types.APIAction{
			Verb:         rbacv1.VerbLaunch,
			ResourceType: rbacv1.ResourceTemplates,
			ResourceName: tmpl.GetName(),
		}
		if EvaluateUser(u, action) {
			filtered = append(filtered, tmpl)
		}
	}
	return filtered
}

// FilterUserNamespaces will take a list of namespaces and filter them based off
// the ones this user can provision desktops in.
func FilterUserNamespaces(u *types.VDIUser, nss []string) []string {
	filtered := make([]string, 0)
	for _, ns := range nss {
		action := &types.APIAction{
			Verb:              rbacv1.VerbLaunch,
			ResourceType:      rbacv1.ResourceTemplates,
			ResourceNamespace: ns,
		}
		if EvaluateUser(u, action) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

// FilterUserServiceAccounts will take a list of service accounts and a given namespace,
// and filter them based off the ones this user can assume with desktops.
func FilterUserServiceAccounts(u *types.VDIUser, sas []string, ns string) []string {
	filtered := make([]string, 0)
	for _, sa := range sas {
		action := &types.APIAction{
			Verb:              rbacv1.VerbUse,
			ResourceType:      rbacv1.ResourceServiceAccounts,
			ResourceName:      sa,
			ResourceNamespace: ns,
		}
		if EvaluateUser(u, action) {
			filtered = append(filtered, sa)
		}
	}
	return filtered
}
