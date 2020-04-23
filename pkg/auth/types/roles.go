package types

import (
	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
)

type Role struct {
	Name             string           `rethinkdb:"id" json:"name"`
	Grants           grants.RoleGrant `rethinkdb:"grants" json:"-"`
	Namespaces       []string         `rethinkdb:"namespaces" json:"namespaces"`
	TemplatePatterns []string         `rethinkdb:"templatePatterns" json:"templatePatterns"`
	GrantNames       []string         `rethinkdb:"-" json:"grants"`
}

func (r *Role) RoleGrants() grants.RoleGrant {
	// in case we want to change how the value gets stored easily
	return r.Grants
}
