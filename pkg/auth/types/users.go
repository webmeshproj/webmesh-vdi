package types

import (
	"regexp"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
)

type User struct {
	Name         string  `rethinkdb:"id" json:"name"`
	Password     string  `rethinkdb:"-" json:"-"`
	PasswordSalt string  `rethinkdb:"password" json:"-"`
	Roles        []*Role `rethinkdb:"role_ids,reference,omitempty" rethinkdb_ref:"id" json:"roles"`
}

func (u *User) HasGrant(grant grants.RoleGrant) bool {
	for _, role := range u.Roles {
		if role.RoleGrants().Has(grant) {
			return true
		}
	}
	return false
}

func (u *User) Namespaces() []string {
	namespaces := make([]string, 0)
	for _, role := range u.Roles {
		for _, restraint := range role.Restraints {
			if restraint.Namespaces != nil && len(restraint.Namespaces) > 0 {
				namespaces = append(namespaces, restraint.Namespaces...)
			}
		}
	}
	return namespaces
}

func (u *User) CanLaunch(namespace, tmpl string) bool {
	if !u.HasGrant(grants.LaunchTemplates) {
		return false
	}
	var namespaceAllowed, tmplAllowed bool

	for _, role := range u.Roles {
		if role.Restraints == nil || len(role.Restraints) == 0 {
			return true
		}
	}

NamespaceLoop:
	for _, role := range u.Roles {
		for _, restraint := range role.Restraints {
			if restraint.Namespaces == nil || len(restraint.Namespaces) == 0 {
				namespaceAllowed = true
				break NamespaceLoop
			}
			if contains(restraint.Namespaces, namespace) {
				namespaceAllowed = true
				break NamespaceLoop
			}
		}
	}

TemplateLoop:
	for _, role := range u.Roles {
		for _, restraint := range role.Restraints {
			if restraint.TemplatePatterns == nil || len(restraint.TemplatePatterns) == 0 {
				tmplAllowed = true
				break TemplateLoop
			}
			for _, pattern := range restraint.TemplatePatterns {
				re := regexp.MustCompile(pattern)
				if re.MatchString(tmpl) {
					tmplAllowed = true
					break TemplateLoop
				}
			}
		}
	}

	return namespaceAllowed && tmplAllowed
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
