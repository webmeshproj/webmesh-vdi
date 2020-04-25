package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// rdbLogger is the logger interface for messagess emitted from rethinkdb functions.
var rdbLogger = logf.Log.WithName("rethinkdb")

// RethinkDBSession is the exported interface for interacting with RethinkDB.
type RethinkDBSession interface {
	// Migrate ensures the required tables and their requested configurations in the database.
	// It also ensures default roles and user accounts.
	Migrate(adminPass string, replicas, shards int32, allowAnonymous bool) error

	// User operations
	GetAllUsers() ([]types.User, error)
	GetUser(id string) (*types.User, error)
	CreateUser(*types.User) error
	UpdateUser(*types.User) error
	SetUserPassword(*types.User, string) error
	DeleteUser(*types.User) error

	// Role operations
	GetAllRoles() ([]*types.Role, error)
	GetRole(string) (*types.Role, error)
	CreateRole(*types.Role) error
	UpdateRole(*types.Role) error
	DeleteRole(*types.Role) error

	// UserSession operations
	GetUserSession(id string) (*types.UserSession, error)
	CreateUserSession(*types.User) (*types.UserSession, error)
	DeleteUserSession(*types.UserSession) error

	Close() error
}

// rethinkDBSession implements the RethinkDBSession interface.
type rethinkDBSession struct {
	// The actual session object. In normal usage, this object contains the connection
	// to rethinkdb. When mocking, this is a mock driver.
	session rdb.QueryExecutor
	// when mocking there is no socket to close
	closeFunc func() error
}

// New returns a RethinkDBSession for the given address. It assumes a
// client TLS configuration is required and present at the expected path.
func New(addr string) (RethinkDBSession, error) {
	tlsConfig, err := tlsutil.NewClientTLSConfig()
	if err != nil {
		return nil, err
	}
	session, err := rdb.Connect(rdb.ConnectOpts{
		Address:   addr,
		TLSConfig: tlsConfig,
	})
	if err != nil {
		return nil, err
	}
	return &rethinkDBSession{
		session:   session,
		closeFunc: func() error { return session.Close() },
	}, nil
}

// NewFromSecret returns a RethinkDBSession at the given address, using the provided
// secret as a source for the TLS client configuration.
func NewFromSecret(c client.Client, addr, name, namespace string) (RethinkDBSession, error) {
	tlsConfig, err := tlsutil.NewClientTLSConfigFromSecret(c, name, namespace)
	if err != nil {
		return nil, err
	}
	session, err := rdb.Connect(rdb.ConnectOpts{
		Address:   addr,
		TLSConfig: tlsConfig,
	})
	if err != nil {
		return nil, err
	}
	return &rethinkDBSession{
		session:   session,
		closeFunc: func() error { return session.Close() },
	}, nil
}

// Close closes a rethinkdb session
func (r *rethinkDBSession) Close() error {
	return r.closeFunc()
}
