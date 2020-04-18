package rethinkdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type UserSession struct {
	ID        string    `rethinkdb:"id" json:"id"`
	ExpiresAt time.Time `rethinkdb:"expires_at" json:"expiresAt"`
	User      *User     `rethinkdb:"user_id,reference" rethinkdb_ref:"id" json:"user"`
}

func (d *rethinkDBSession) GetUserSession(id string) (*UserSession, error) {
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
	session := &UserSession{}
	if err := cursorIntoObj(cursor, session); err != nil {
		return nil, err
	}
	for _, role := range session.User.Roles {
		role.GrantNames = role.Grants.Names()
	}
	return session, nil
}

func (d *rethinkDBSession) CreateUserSession(user *User) (*UserSession, error) {
	session := &UserSession{
		ID:        uuid.New().String(),
		ExpiresAt: time.Now().Add(DefaultSessionLength),
		User:      user,
	}
	for _, role := range session.User.Roles {
		role.GrantNames = role.Grants.Names()
	}
	cursor, err := rdb.DB(kvdiDB).Table(userSessionTable).Insert(session).Run(d.session)
	if err != nil {
		return nil, err
	}
	return session, cursor.Err()
}
