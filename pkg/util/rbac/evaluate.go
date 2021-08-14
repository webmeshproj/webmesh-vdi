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
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/types"
)

// EvaluateUser will iterate the user's roles and return true if any of them have
// a rule that allows the given action.
func EvaluateUser(u *types.VDIUser, action *types.APIAction) bool {
	for _, role := range u.Roles {
		if ok := EvaluateRole(role, action); ok {
			return true
		}
	}
	return false
}

// EvaluateRole iterates all the rules in the given role role and returns true if any of them
// allow the provided action.
func EvaluateRole(r *types.VDIUserRole, action *types.APIAction) bool {
	for _, rule := range r.Rules {
		if ok := EvaluateRule(rule, action); ok {
			return true
		}
	}
	return false
}

// EvaluateRule checks if the given rule allows the given action. First the verb is matched,
// then the resource type, and then optionally a name and namespace.
func EvaluateRule(r rbacv1.Rule, action *types.APIAction) bool {
	if action.ResourceType == rbacv1.ResourceServiceAccounts && action.ResourceName == "default" {
		// Treat default service accounts as just checking the ability to launch templates
		// in the given namespace
		action.ResourceName = ""
		action.ResourceType = rbacv1.ResourceTemplates
	}
	if !r.HasVerb(action.Verb) {
		return false
	}
	if !r.HasResourceType(action.ResourceType) {
		return false
	}
	if action.ResourceName != "" && !r.MatchesResourceName(action.ResourceName) {
		return false
	}
	if action.ResourceNamespace != "" && !r.HasNamespace(action.ResourceNamespace) {
		return false
	}
	return true
}
