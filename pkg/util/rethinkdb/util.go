package rethinkdb

import (
	"fmt"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
)

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func RDBAddrForCR(cr *v1alpha1.VDICluster) string {
	return fmt.Sprintf("%s.%s.svc:%d", cr.GetRethinkDBProxyName(), cr.GetCoreNamespace(), v1alpha1.RethinkDBDriverPort)
}

func cursorIntoObj(cursor *rdb.Cursor, obj interface{}) error {
	if err := cursor.One(obj); err != nil {
		return err
	} else if cursor.Err() != nil {
		return cursor.Err()
	}
	return nil
}

func cursorIntoObjSlice(cursor *rdb.Cursor, obj interface{}) error {
	if err := cursor.All(obj); err != nil {
		return err
	} else if cursor.Err() != nil {
		return cursor.Err()
	}
	return nil
}
