package rethinkdb

import (
	"fmt"

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
