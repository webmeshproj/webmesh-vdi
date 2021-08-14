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

// UserIncludesRule returns true if the rules applied to this user are not elevated
// by any of the permissions in the provided rule.
func UserIncludesRule(u *types.VDIUser, ruleToCheck rbacv1.Rule, resourceGetter types.ResourceGetter) bool {
	for _, role := range u.Roles {
		if ok := RoleIncludesRule(role, ruleToCheck, resourceGetter); ok {
			return true
		}
	}
	return false
}

// RoleIncludesRule returns true if the rules applied to this role are not elevated
// by any of the permissions in the provided rule.
func RoleIncludesRule(r *types.VDIUserRole, ruleToCheck rbacv1.Rule, resourceGetter types.ResourceGetter) bool {
	for _, rule := range r.Rules {
		if ok := RuleIncludes(rule, ruleToCheck, resourceGetter); ok {
			return true
		}
	}
	return false
}

// RuleIncludes returns false if ruleToCheck matches any actions or resources that r does not.
func RuleIncludes(r, ruleToCheck rbacv1.Rule, resourceGetter types.ResourceGetter) bool {

	if r.DeepEqual(ruleToCheck) {
		return true
	}

	for _, verb := range ruleToCheck.Verbs {
		if !r.HasVerb(verb) {
			return false
		}
	}
	for _, ns := range ruleToCheck.Namespaces {
		if !r.HasNamespace(ns) {
			return false
		}
	}
	for _, resource := range ruleToCheck.Resources {
		if !r.HasResourceType(resource) {
			return false
		}
		// If any of the functions below fail it will be important for the caller
		// defining the Getters to log the error appropriately. The request will get
		// denied and that will be the only way to see the actual error.

		// There are some caveats to supporting regex patterns for resources. The
		// easiest way to pseudo check equality is to just grab a list of currently
		// available resources and see if the proposed rule matches them but ours doesn't.
		// There are holes, because a user could create a regex that matches a future
		// resource that their current rule does not.
		if resource == rbacv1.ResourceAll || resource == rbacv1.ResourceRoles {
			roles, err := resourceGetter.GetRoles()
			if err != nil {
				return false
			}
			for _, role := range roles {
				if ruleToCheck.MatchesResourceName(role.GetName()) && !r.MatchesResourceName(role.GetName()) {
					return false
				}
			}
		}
		if resource == rbacv1.ResourceAll || resource == rbacv1.ResourceUsers {
			users, err := resourceGetter.GetUsers()
			if err != nil {
				return false
			}
			for _, user := range users {
				if ruleToCheck.MatchesResourceName(user.GetName()) && !r.MatchesResourceName(user.GetName()) {
					return false
				}
			}
		}
		if resource == rbacv1.ResourceAll || resource == rbacv1.ResourceTemplates {
			templates, err := resourceGetter.GetTemplates()
			if err != nil {
				return false
			}
			for _, tmpl := range templates {
				if ruleToCheck.MatchesResourceName(tmpl) && !r.MatchesResourceName(tmpl) {
					return false
				}
			}
		}
	}
	return true
}
