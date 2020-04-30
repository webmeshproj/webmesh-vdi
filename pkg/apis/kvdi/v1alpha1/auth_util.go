package v1alpha1

import (
	"fmt"
	"reflect"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetAdminRole returns an admin role for this VDICluster.
func (v *VDICluster) GetAdminRole() *VDIRole {
	return &VDIRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-admin", v.GetName()),
			Labels: map[string]string{
				RoleClusterRefLabel: v.GetName(),
			},
		},
		Rules: []Rule{
			{
				Verbs:            []Verb{VerbAll},
				Resources:        []Resource{ResourceAll},
				ResourcePatterns: []string{".*"},
				Namespaces:       []string{NamespaceAll},
			},
		},
	}
}

// GetName returns the name of a VDIUser.
func (u *VDIUser) GetName() string { return u.Name }

// Evaluate will iterate the user's roles and return true if any of them have
// a rule that allows the given action.
func (u *VDIUser) Evaluate(action *APIAction) bool {
	for _, role := range u.Roles {
		if ok := role.Evaluate(action); ok {
			return true
		}
	}
	return false
}

// IncludesRule returns true if the rules applied to this user are not elevated
// by any of the permissions in the provided rule.
func (u *VDIUser) IncludesRule(ruleToCheck Rule, resourceGetter ResourceGetter) bool {
	for _, role := range u.Roles {
		if ok := role.IncludesRule(ruleToCheck, resourceGetter); !ok {
			return false
		}
	}
	return true
}

// FilterNamespaces will take a list of namespaces, and filter them based off
// the ones this user can provision desktops in.
func (u *VDIUser) FilterNamespaces(nss []string) []string {
	filtered := make([]string, 0)
	for _, ns := range nss {
		action := &APIAction{
			Verb:              VerbLaunch,
			ResourceType:      ResourceTemplates,
			ResourceNamespace: ns,
		}
		if u.Evaluate(action) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

// FilterTemplates will take a list of DesktopTemplates and filter them based
// off which ones the user is allowed to use.
func (u *VDIUser) FilterTemplates(tmpls []DesktopTemplate) []DesktopTemplate {
	filtered := make([]DesktopTemplate, 0)
	for _, tmpl := range tmpls {
		action := &APIAction{
			Verb:         VerbLaunch,
			ResourceType: ResourceTemplates,
			ResourceName: tmpl.GetName(),
		}
		if u.Evaluate(action) {
			filtered = append(filtered, tmpl)
		}
	}
	return filtered
}

// Evaluate iterates all the rules in this role and returns true if any of them
// allow the provided action.
func (r *VDIUserRole) Evaluate(action *APIAction) bool {
	for _, rule := range r.Rules {
		if ok := rule.Evaluate(action); ok {
			return true
		}
	}
	return false
}

// IncludesRule returns true if the rules applied to this role are not elevated
// by any of the permissions in the provided rule.
func (r *VDIUserRole) IncludesRule(ruleToCheck Rule, resourceGetter ResourceGetter) bool {
	for _, rule := range r.Rules {
		if ok := rule.IncludesRule(ruleToCheck, resourceGetter); !ok {
			return false
		}
	}
	return true
}

// Evaluate checks if this rule allows the given action. First the verb is matched,
// then the resource type, and then optionally a name and namespace.
func (r *Rule) Evaluate(action *APIAction) bool {
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
				if ruleToCheck.MatchesResourceName(tmpl.GetName()) && !r.MatchesResourceName(tmpl.GetName()) {
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
