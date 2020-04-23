package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var rdbLogger = logf.Log.WithName("rethinkdb")

type RethinkDBSession interface {
	Migrate(adminPass string, replicas, shards int32, allowAnonymous bool) error
	GetAllUsers() ([]types.User, error)
	GetUser(id string) (*types.User, error)
	CreateUser(*types.User) error
	SetUserPassword(*types.User, string) error
	DeleteUser(*types.User) error
	GetAllRoles() ([]*types.Role, error)
	GetRole(string) (*types.Role, error)
	CreateRole(*types.Role) error
	GetUserSession(id string) (*types.UserSession, error)
	CreateUserSession(*types.User) (*types.UserSession, error)
	DeleteUserSession(*types.UserSession) error
	Close() error
}

type rethinkDBSession struct {
	session *rdb.Session
}

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
	return &rethinkDBSession{session: session}, nil
}

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
	return &rethinkDBSession{session: session}, nil
}

func (r *rethinkDBSession) Close() error {
	return r.session.Close()
}
