package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/grants"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type Role struct {
	Name       string           `rethinkdb:"id" json:"name"`
	Grants     grants.RoleGrant `rethinkdb:"grants" json:"-"`
	GrantNames []string         `rethinkdb:"-" json:"grants"`
}

func (d *rethinkDBSession) GetRole(name string) (*Role, error) {
	cursor, err := rdb.DB(kvdiDB).Table(rolesTable).Get(name).Run(d.session)
	if err != nil {
		return nil, err
	}
	if cursor.IsNil() {
		return nil, errors.NewRoleNotFoundError(name)
	}
	role := &Role{}
	role.GrantNames = role.Grants.Names()
	return role, cursorIntoObj(cursor, role)
}

func (d *rethinkDBSession) CreateRole(role *Role) error {
	cursor, err := rdb.DB(kvdiDB).Table(rolesTable).Insert(role).Run(d.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}
