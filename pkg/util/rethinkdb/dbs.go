package rethinkdb

import (
	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func (r *rethinkDBSession) listDBs() ([]string, error) {
	cursor, err := rdb.DBList().Run(r.session)
	if err != nil {
		return nil, err
	}
	dbs := make([]string, 0)
	if err := cursor.All(&dbs); err != nil {
		return nil, err
	}
	return dbs, cursor.Err()
}

func (r *rethinkDBSession) listDBTables(db string) ([]string, error) {
	cursor, err := rdb.DB(db).TableList().Run(r.session)
	if err != nil {
		return nil, err
	}
	var tables []string
	if err = cursor.All(&tables); err != nil {
		return nil, err
	}
	return tables, cursor.Err()
}

func (r *rethinkDBSession) createDB(name string) error {
	cursor, err := rdb.DBCreate(name).Run(r.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}

func (r *rethinkDBSession) deleteDB(name string) error {
	cursor, err := rdb.DBDrop(name).Run(r.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}
