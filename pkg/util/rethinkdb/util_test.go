package rethinkdb

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
)

func TestRDBAddrForCR(t *testing.T) {
	cluster := &v1alpha1.VDICluster{}
	cluster.Name = "test-cluster"
	if name := RDBAddrForCR(cluster); name != "test-cluster-rethinkdb-proxy.default.svc:28015" {
		t.Error("Got unexpected rethinkdb addr:", name)
	}
}
