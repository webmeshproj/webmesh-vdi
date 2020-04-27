package rethinkdb

import (
	"errors"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

// some test valuse
var nonExist = "non-exist"
var newItem = "new-item"
var errItem = "err-item"
var testHash = "test-hash"

// External test values
var SuccessItem = "success"
var ErrorItem = "error"

// mockQueries contains the queries the mock session gets configured to serve.
// Essentially, this is a declaration of the state of the mock database.
var mockQueries = []struct {
	Query  rdb.Term
	Result interface{}
	Error  error
}{
	{
		Query:  rdb.DBList(),
		Result: []string{kvdiDB},
	},
	{
		Query: rdb.DB(kvdiDB).TableList(),
		Result: []string{
			string(usersTable),
			string(rolesTable),
		},
	},
	{
		Query:  rdb.DB(kvdiDB).Table(usersTable).Config().Field("shards").Count(),
		Result: 1,
	},
	{
		Query:  rdb.DB(kvdiDB).Table(rolesTable).Config().Field("shards").Count(),
		Result: 1,
	},
	{
		Query:  rdb.DB(kvdiDB).Table(usersTable).Config().Field("shards").Nth(0).Field("replicas").Count(),
		Result: 1,
	},
	{
		Query:  rdb.DB(kvdiDB).Table(rolesTable).Config().Field("shards").Nth(0).Field("replicas").Count(),
		Result: 1,
	},
	{
		Query:  rdb.DB(kvdiDB).Table(usersTable).Get(adminUser),
		Result: map[string]interface{}{"id": adminUser, "role_ids": []string{adminRole}, "password": testHash},
	},
	{
		Query:  rdb.DB(kvdiDB).Table(rolesTable).Get(adminRole),
		Result: map[string]interface{}{"id": adminRole, "grants": grants.All},
	},
	{
		Query:  rdb.DB(kvdiDB).Table(rolesTable).Get(launchTemplateRole),
		Result: map[string]interface{}{"id": launchTemplateRole, "grants": grants.LaunchTemplatesGrant},
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Map(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Result: []interface{}{
			map[string]interface{}{
				"id":       adminUser,
				"password": testHash,
				"roles": map[string]interface{}{
					"name":             adminRole,
					"namespaces":       []string{},
					"templatePatterns": []string{},
					"grants":           grants.All,
				},
			},
			map[string]interface{}{
				"id":       anonymousUser,
				"password": testHash,
				"roles": map[string]interface{}{
					"name":             launchTemplateRole,
					"namespaces":       []string{},
					"templatePatterns": []string{},
					"grants":           grants.LaunchTemplatesGrant,
				},
			},
		},
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(adminUser).Do(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Result: map[string]interface{}{
			"id":       adminUser,
			"password": testHash,
			"roles": map[string]interface{}{
				"name":             adminRole,
				"namespaces":       []string{},
				"templatePatterns": []string{},
				"grants":           grants.All,
			},
		},
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(anonymousUser).Do(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Result: map[string]interface{}{
			"id":       anonymousUser,
			"password": testHash,
			"roles": map[string]interface{}{
				"name":             launchTemplateRole,
				"namespaces":       []string{},
				"templatePatterns": []string{},
				"grants":           grants.LaunchTemplatesGrant,
			},
		},
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(nonExist).Do(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Result: nil,
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(SuccessItem).Do(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Result: map[string]interface{}{
			"id":       SuccessItem,
			"password": testHash,
		},
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(ErrorItem).Do(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(errItem).Do(func(row rdb.Term) interface{} {
			return rdb.Branch(row, row.Merge(func(plan rdb.Term) interface{} {
				return map[string]interface{}{
					"role_ids": rdb.DB(kvdiDB).Table(rolesTable).GetAll(rdb.Args(plan.Field("role_ids"))).CoerceTo("array"),
				}
			}), nil)
		}),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(adminUser).Update(map[string]interface{}{
			"password": testHash,
		}),
		Result: map[string]interface{}{},
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable),
		Result: []interface{}{
			map[string]interface{}{
				"id":     adminRole,
				"grants": grants.All,
			},
			map[string]interface{}{
				"id":     launchTemplateRole,
				"grants": grants.LaunchTemplatesGrant,
			},
		},
	},
	{
		Query:  rdb.DB(kvdiDB).Table(rolesTable).Get(nonExist),
		Result: nil,
	},
	{
		Query:  rdb.DB(kvdiDB).Table(usersTable).Get(nonExist),
		Result: nil,
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Get(errItem),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(errItem),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Insert(&types.Role{
			Name: newItem,
		}),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Insert(&types.User{
			Name:         newItem,
			PasswordSalt: testHash,
		}),
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Insert(&types.Role{Name: errItem}),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Insert(&types.User{Name: errItem, PasswordSalt: testHash}),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Get(newItem).Update(&types.Role{
			Name: newItem,
		}),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(newItem).Update(map[string]interface{}{
			"role_ids": []string{newItem},
		}),
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Get(errItem).Update(&types.Role{Name: errItem}),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(errItem).Update(map[string]interface{}{
			"role_ids": []string{errItem},
		}),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Get(newItem).Delete(),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(newItem).Delete(),
	},
	{
		Query: rdb.DB(kvdiDB).Table(rolesTable).Get(errItem).Delete(),
		Error: errors.New(""),
	},
	{
		Query: rdb.DB(kvdiDB).Table(usersTable).Get(errItem).Delete(),
		Error: errors.New(""),
	},
}

// NewMock returns a session with a mock rethinkdb driver that is pre-populated
// with some expected queries.
func NewMock(args ...interface{}) RethinkDBSession {
	hashFunc = func(string) (string, error) { return testHash, nil }

	mock := rdb.NewMock()

	for _, query := range mockQueries {
		mock.On(query.Query).Return(query.Result, query.Error)
	}

	return &rethinkDBSession{
		session:   mock,
		closeFunc: func() error { return nil },
	}
}
