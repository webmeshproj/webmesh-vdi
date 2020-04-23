package types

import (
	"regexp"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
)

type User struct {
	Name         string  `rethinkdb:"id" json:"name"`
	Password     string  `rethinkdb:"-" json:"-"`
	PasswordSalt string  `rethinkdb:"password" json:"-"`
	Roles        []*Role `rethinkdb:"role_ids,reference" rethinkdb_ref:"id" json:"roles"`
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
		if role.Namespaces != nil && len(role.Namespaces) > 0 {
			namespaces = append(namespaces, role.Namespaces...)
		}
	}
	return namespaces
}

func (u *User) CanLaunch(namespace, tmpl string) bool {
	if !u.HasGrant(grants.LaunchTemplates) {
		return false
	}
	var namespaceAllowed, tmplAllowed bool

NamespaceLoop:
	for _, role := range u.Roles {
		if role.Namespaces == nil || len(role.Namespaces) == 0 {
			namespaceAllowed = true
			break NamespaceLoop
		}
		if contains(role.Namespaces, namespace) {
			namespaceAllowed = true
			break NamespaceLoop
		}
	}

TemplateLoop:
	for _, role := range u.Roles {
		if role.TemplatePatterns == nil || len(role.TemplatePatterns) == 0 {
			tmplAllowed = true
			break TemplateLoop
		}
		for _, pattern := range role.TemplatePatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(tmpl) {
				tmplAllowed = true
				break TemplateLoop
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
