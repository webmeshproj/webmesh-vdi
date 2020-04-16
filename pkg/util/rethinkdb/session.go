package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RethinkDBSession interface {
	Migrate(replicas, shards int32) error
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
