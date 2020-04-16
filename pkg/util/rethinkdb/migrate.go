package rethinkdb

import (
	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

const (
	kvdiDB = "kvdi"

	userTable            = "users"
	userSessionTable     = "userSessions"
	desktopSessionsTable = "desktopSessions"
)

var allTables = []string{userTable, userSessionTable, desktopSessionsTable}

func (r *rethinkDBSession) Migrate(desiredReplicas, desiredShards int32) error {
	// Setup DBs
	dbs, err := r.listDBs()
	if err != nil {
		return err
	}
	if contains(dbs, "test") {
		if err := r.deleteDB("test"); err != nil {
			return err
		}
	}
	if !contains(dbs, kvdiDB) {
		if err := r.createDB(kvdiDB); err != nil {
			return err
		}
	}

	// Setup Tables
	tables, err := r.listDBTables(kvdiDB)
	if err != nil {
		return err
	}
	for _, table := range allTables {
		if !contains(tables, table) {
			if err := r.createTable(kvdiDB, table); err != nil {
				return err
			}
		}
		shards, replicas, err := r.getTableConfig(kvdiDB, table)
		if err != nil {
			return err
		}
		if replicas != desiredReplicas || shards != desiredShards {
			if cursor, err := rdb.DB(kvdiDB).Table(table).Reconfigure(rdb.ReconfigureOpts{
				Replicas: desiredReplicas,
				Shards:   desiredShards,
			}).Run(r.session); err != nil {
				return err
			} else if cursor.Err() != nil {
				return cursor.Err()
			}
		}
	}

	return nil
}
