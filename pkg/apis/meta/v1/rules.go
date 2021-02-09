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

package v1

import (
	"reflect"
	"regexp"
)

// Rule represents a set of permissions applied to a VDIRole. It mostly resembles
// an rbacv1.PolicyRule, with resources being a regex and the addition of a
// namespace selector.
type Rule struct {
	// The actions this rule applies for. VerbAll matches all actions.
	Verbs []Verb `json:"verbs,omitempty"`
	// Resources this rule applies to. ResourceAll matches all resources.
	Resources []Resource `json:"resources,omitempty"`
	// Resource regexes that match this rule. This can be template patterns, role
	// names or user names. There is no All representation because * will have
	// that effect on its own when the regex is evaluated. When referring to "serviceaccounts",
	// only the "use" verb is evaluated in the context of assuming those accounts in
	// desktop sessions.
	//
	// **NOTE**: The `kvdi-manager` is responsible for launching pods with a service account
	// requested for a given Desktop. If the service account itself contains more permissions
	// than the manager itself, the Kubernetes API will deny the request. The way to remedy this
	// would be to either mirror permissions to that ClusterRole, or make the `kvdi-manager` itself a
	// cluster admin, both of which come with inherent risks. In the end, you can decide the best
	// approach for your use case with regards to exposing access to the Kubernetes APIs via kvdi sessions.
	ResourcePatterns []string `json:"resourcePatterns,omitempty"`
	// Namespaces this rule applies to. Only evaluated for template launching
	// permissions. NamespaceAll matches all namespaces.
	Namespaces []string `json:"namespaces,omitempty"`
}

// Evaluate checks if this rule allows the given action. First the verb is matched,
// then the resource type, and then optionally a name and namespace.
func (r *Rule) Evaluate(action *APIAction) bool {
	if action.ResourceType == ResourceServiceAccounts && action.ResourceName == "default" {
		// Treat default service accounts as just checking the ability to launch templates
		// in the given namespace
		action.ResourceName = ""
		action.ResourceType = ResourceTemplates
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

// DeepEqual returns true if the provided rule matches this one exactly.
func (r *Rule) DeepEqual(rule Rule) bool {
	return reflect.DeepEqual(r.Verbs, rule.Verbs) &&
		reflect.DeepEqual(r.Resources, rule.Resources) &&
		reflect.DeepEqual(r.ResourcePatterns, rule.ResourcePatterns) &&
		reflect.DeepEqual(r.Namespaces, rule.Namespaces)
}

// IncludesRule returns false if the given rule matches any actions or resources
// that this rule does not.
func (r *Rule) IncludesRule(ruleToCheck Rule, resourceGetter ResourceGetter) bool {

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
		if resource == ResourceAll || resource == ResourceRoles {
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
		if resource == ResourceAll || resource == ResourceUsers {
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
		if resource == ResourceAll || resource == ResourceTemplates {
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

// HasVerb returns true if this rule contains the given verb.
func (r *Rule) HasVerb(verb Verb) bool {
	for _, item := range r.Verbs {
		if item == VerbAll {
			return true
		}
		if item == verb {
			return true
		}
	}
	return false
}

// HasResourceType returns true if this rule has the given resource type.
func (r *Rule) HasResourceType(resource Resource) bool {
	for _, item := range r.Resources {
		if item == ResourceAll {
			return true
		}
		if item == resource {
			return true
		}
	}
	return false
}

// MatchesResourceName returns true if any of the resource patterns in this rule
// match the given name.
func (r *Rule) MatchesResourceName(name string) bool {
	for _, pattern := range r.ResourcePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Should have an external validator to let the user know
			// there is a bad regex.
			continue
		}
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

// HasNamespace returns true if this rule includes the given namespace.
func (r *Rule) HasNamespace(ns string) bool {
	for _, item := range r.Namespaces {
		if item == NamespaceAll {
			return true
		}
		if item == ns {
			return true
		}
	}
	return false
}
