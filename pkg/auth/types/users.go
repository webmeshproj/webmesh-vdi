package types

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

// User represents a kVDI user. A reference to a role and the password salt
// are stored in the database. A user is allowed access to a resource if any
// of their attached roles allows that access.
type User struct {
	Name         string  `rethinkdb:"id" json:"name"`
	Password     string  `rethinkdb:"-" json:"-"`
	PasswordSalt string  `rethinkdb:"password" json:"-"`
	Roles        []*Role `rethinkdb:"role_ids,reference" rethinkdb_ref:"id" json:"roles"`
}

// HasGrant returns true if any of the user's roles contains the requested grant.
func (u *User) HasGrant(grant grants.RoleGrant) bool {
	for _, role := range u.Roles {
		if role.RoleGrants().Has(grant) {
			return true
		}
	}
	return false
}

// FilterNamespaces takes a list of namespaces and filters it based on the ones
// the user is allowed to launch templates in.
func (u *User) FilterNamespaces(nss []string) []string {
	namespaces := make([]string, 0)
	for _, role := range u.Roles {
		if len(role.Namespaces) == 0 {
			return nss
		}
		for _, ns := range role.Namespaces {
			if common.StringSliceContains(nss, ns) {
				namespaces = common.AppendStringIfMissing(namespaces, ns)
			}
		}
	}
	return namespaces
}

// FilterTemplates takes a list of templates and filters it based on the ones
// the user is allowed to launch.
func (u *User) FilterTemplates(tmpls []v1alpha1.DesktopTemplate) []v1alpha1.DesktopTemplate {
	filtered := make([]v1alpha1.DesktopTemplate, 0)
TmplLoop:
	for _, tmpl := range tmpls {
		for _, role := range u.Roles {
			if role.MatchesTemplatePattern(tmpl.GetName()) {
				filtered = append(filtered, tmpl)
				continue TmplLoop
			}
		}
	}
	return filtered
}

// RoleNames returns a list of the role names for this user.
func (u *User) RoleNames() []string {
	roles := make([]string, 0)
	for _, role := range u.Roles {
		roles = append(roles, role.Name)
	}
	return roles
}

// CanLaunch returns true if any of the user's roles allows launching the
// requested template in the requested namespace.
func (u *User) CanLaunch(namespace, tmpl string) bool {
	if !u.HasGrant(grants.LaunchTemplates) {
		return false
	}
	for _, role := range u.Roles {
		if role.CanLaunch(namespace, tmpl) {
			return true
		}
	}
	return false
}

// ElevatedBy returns true if any of the permissions in the provided role
// would be an elevation of privileges for this user.
func (u *User) ElevatedBy(role *Role) bool {
	// If this user doesn't have the grant at all, assume true
	for _, grant := range role.Grants.Grants() {
		if !u.HasGrant(grant) {
			return true
		}
	}

	// Only test launch conditions if we are considering LaunchTemplate access
	if !role.Grants.Has(grants.LaunchTemplates) {
		return false
	}

	// For each user role, evaluate if there are any new namespaces or patterns
	// in the provided role.
	for _, userRole := range u.Roles {
		if !userRole.MatchesNamespaces(role.Namespaces) {
			return true
		}
		for _, pattern := range role.TemplatePatterns {
			if !userRole.HasTemplatePattern(pattern) {
				return true
			}
		}
	}

	// All checks passed - not an elevation
	return false
}
