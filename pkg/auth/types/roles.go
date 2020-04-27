package types

import (
	"regexp"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

// Role represents a set of permissions that can be attached to a kVDI user.
// Currently there are three main components:
//   - `Grants`: The 'verb' permissions of this role
//   - `Namespaces`: Kubernetes namespaces where the role can launch templates
//   - `Template Patterns`: Specific patterns that the template trying to be launched
//      must match.
type Role struct {
	Name             string           `rethinkdb:"id" json:"name"`
	Grants           grants.RoleGrant `rethinkdb:"grants" json:"grantBit"`
	Namespaces       []string         `rethinkdb:"namespaces" json:"namespaces"`
	TemplatePatterns []string         `rethinkdb:"templatePatterns" json:"templatePatterns"`
	GrantNames       []string         `rethinkdb:"-" json:"grants"`
}

// RoleGrants returns the numerical grant value for the role.
func (r *Role) RoleGrants() grants.RoleGrant {
	// in case we want to change how the value gets stored easily
	return r.Grants
}

// MatchesNamespaces returns true if this role's allowed namespaces grant
// permission to all the provided ones.
func (r *Role) MatchesNamespaces(nss []string) bool {
	if len(nss) == 0 {
		return len(r.Namespaces) == 0
	}
	for _, ns := range nss {
		if r.HasNamespace(ns) {
			return true
		}
	}
	return false
}

// HasNamespace returns true if this role is allowed to use the given namespace.
func (r *Role) HasNamespace(namespace string) bool {
	if len(r.Namespaces) == 0 {
		return true
	}
	return common.StringSliceContains(r.Namespaces, namespace)
}

// HasTemplatePattern returns true if this role has a temeplate pattern that
// includes the provided one. Note this doesn't actually evaluate the pattern,
// just checks that it isn't a new one.
func (r *Role) HasTemplatePattern(pattern string) bool {
	if len(r.TemplatePatterns) == 0 {
		return true
	}
	return common.StringSliceContains(r.TemplatePatterns, pattern)
}

// MatchesTemplatePattern returns true if the given template matches one of the
// patterns this role has access to.
func (r *Role) MatchesTemplatePattern(tmpl string) bool {
	if len(r.TemplatePatterns) == 0 {
		return true
	}
	// Trust that the regex was validated by the API before being
	// written to the database. So I guess also trust that the user isn't
	// messing with the database directly.
	for _, pattern := range r.TemplatePatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(tmpl) {
			return true
		}
	}

	return false
}

// CanLaunch returns true if this role has the ability to launch a desktop
// in the given namespace with the given template.
func (r *Role) CanLaunch(namespace, tmpl string) bool {
	if !r.RoleGrants().Has(grants.LaunchTemplates) {
		return false
	}
	return r.HasNamespace(namespace) && r.MatchesTemplatePattern(tmpl)
}
