package rethinkdb

import (
	"reflect"
	"testing"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func TestListDBs(t *testing.T) {
	mock := rdb.NewMock()
	sess := &rethinkDBSession{session: mock}
	mock.On(rdb.DBList()).Return([]interface{}{
		string("test"),
	}, nil)
	dbs, err := sess.listDBs()
	if err != nil {
		t.Error("Expected no error, got:", err)
	} else if !reflect.DeepEqual(dbs, []string{"test"}) {
		t.Error("Did not get expected db list")
	}
}

func TestListDBTables(t *testing.T) {
	mock := rdb.NewMock()
	sess := &rethinkDBSession{session: mock}
	mock.On(rdb.DB("test").TableList()).Return([]interface{}{
		string("test"),
	}, nil)
	tables, err := sess.listDBTables("test")
	if err != nil {
		t.Error("Expected no error, got:", err)
	} else if !reflect.DeepEqual(tables, []string{"test"}) {
		t.Error("Did not get expected table list")
	}
}

func TestCreateDB(t *testing.T) {
	mock := rdb.NewMock()
	mock.On(rdb.DBCreate("test")).Return([]interface{}{}, nil)
	sess := &rethinkDBSession{session: mock}
	if err := sess.createDB("test"); err != nil {
		t.Error(err)
	}
}

func TestDeleteDB(t *testing.T) {
	mock := rdb.NewMock()
	mock.On(rdb.DBDrop("test")).Return([]interface{}{}, nil)
	sess := &rethinkDBSession{session: mock}
	if err := sess.deleteDB("test"); err != nil {
		t.Error(err)
	}
}
