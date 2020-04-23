package rethinkdb

import (
	"time"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/google/uuid"
	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func (d *rethinkDBSession) GetUserSession(id string) (*types.UserSession, error) {
	cursor, err := rdb.DB(kvdiDB).Table(userSessionTable).Get(id).Do(func(row rdb.Term) interface{} {
		return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
			return map[string]interface{}{
				"user_id": rdb.DB(kvdiDB).Table(usersTable).Get(plan.Field("user_id")).Do(
					func(innerrow rdb.Term) interface{} {
						return rdb.Branch(innerrow, innerrow.Merge(func(innerplan rdb.Term) interface{} {
							return map[string]interface{}{
								"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(innerplan.Field("role_ids"))).CoerceTo("array"),
							}
						}), nil)
					}),
			}
		}), nil)
	}).Run(d.session)
	if err != nil {
		return nil, err
	}
	if cursor.IsNil() {
		return nil, errors.NewUserSessionNotFoundError(id)
	}
	session := &types.UserSession{}
	if err := cursorIntoObj(cursor, session); err != nil {
		return nil, err
	}
	for _, role := range session.User.Roles {
		role.GrantNames = role.RoleGrants().Names()
	}
	return session, nil
}

func (d *rethinkDBSession) CreateUserSession(user *types.User) (*types.UserSession, error) {
	session := &types.UserSession{
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(DefaultSessionLength),
		User:      user,
	}
	for _, role := range session.User.Roles {
		role.GrantNames = role.RoleGrants().Names()
	}
	cursor, err := rdb.DB(kvdiDB).Table(userSessionTable).Insert(session).Run(d.session)
	if err != nil {
		return nil, err
	}
	return session, cursor.Err()
}

func (d *rethinkDBSession) DeleteUserSession(session *types.UserSession) error {
	cursor, err := rdb.DB(kvdiDB).Table(userSessionTable).Get(session.Token).Delete().Run(d.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}
