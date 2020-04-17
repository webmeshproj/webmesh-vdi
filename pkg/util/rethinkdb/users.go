package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type User struct {
	Name         string `rethinkdb:"id" json:"name"`
	Password     string `rethinkdb:"-" json:"-"`
	PasswordSalt string `rethinkdb:"password" json:"-"`
	Roles        []Role `rethinkdb:"role_ids,reference,omitempty" rethinkdb_ref:"id" json:"roles"`
}

func (d *rethinkDBSession) GetAllUsers() ([]User, error) {
	cursor, err := rdb.DB(kvdiDB).Table(usersTable).ForEach(func(row rdb.Term) interface{} {
		return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
			return map[string]interface{}{
				"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
			}
		}), nil)
	}).Run(d.session)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0)
	if cursor.IsNil() {
		return users, nil
	}
	return users, cursorIntoObjSlice(cursor, &users)
}

func (d *rethinkDBSession) GetUser(name string) (*User, error) {
	cursor, err := rdb.DB(kvdiDB).Table(usersTable).Get(name).Do(func(row rdb.Term) interface{} {
		return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
			return map[string]interface{}{
				"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
			}
		}), nil)
	}).Run(d.session)
	if err != nil {
		return nil, err
	}
	if cursor.IsNil() {
		return nil, errors.NewUserNotFoundError(name)
	}
	user := &User{}
	return user, cursorIntoObj(cursor, user)
}

func (d *rethinkDBSession) CreateUser(user *User) error {
	hash, err := util.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.PasswordSalt = string(hash)
	cursor, err := rdb.DB(kvdiDB).Table(usersTable).Insert(user).Run(d.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}

func (d *rethinkDBSession) SetUserPassword(user *User, password string) error {
	hash, err := util.HashPassword(password)
	if err != nil {
		return err
	}
	user.PasswordSalt = string(hash)
	cursor, err := rdb.DB(kvdiDB).Table(usersTable).Get(user.Name).Update(user).Run(d.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}
