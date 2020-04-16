package rethinkdb

import (
	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func (r *rethinkDBSession) createTable(db, table string) error {
	cursor, err := rdb.DB(db).TableCreate(table).Run(r.session)
	if err != nil {
		return err
	}
	return cursor.Err()
}

func (r *rethinkDBSession) getTableConfig(db, table string) (shards, replicas int32, err error) {
	var res *rdb.Cursor
	res, err = rdb.DB(db).
		Table(table).
		Config().
		Field("shards").
		Nth(0).
		Field("replicas").
		Count().
		Run(r.session)
	if err != nil {
		return
	} else if res.Err() != nil {
		err = res.Err()
		return
	}
	if err = res.One(&replicas); err != nil {
		return
	}
	res, err = rdb.DB(db).
		Table(table).
		Config().
		Field("shards").
		Count().
		Run(r.session)
	if err != nil {
		return
	} else if res.Err() != nil {
		err = res.Err()
		return
	}
	err = res.One(&shards)
	return
}
