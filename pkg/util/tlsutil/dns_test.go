package tlsutil

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

func TestDNSNames(t *testing.T) {
	dnsNames := DNSNames("test-service", "test-namespace")
	expectedDNSNames := []string{
		"test-service",
		"test-service.test-namespace",
		"test-service.test-namespace.svc",
		fmt.Sprintf("test-service.test-namespace.svc.%s", common.GetClusterSuffix()),
	}
	if !reflect.DeepEqual(dnsNames, expectedDNSNames) {
		t.Error(dnsNames)
	}
}

func TestHeadlessDNSNames(t *testing.T) {
	dnsNames := HeadlessDNSNames("test-pod", "test-service", "test-namespace")
	expectedDNSNames := []string{
		"test-service",
		"test-service.test-namespace",
		"test-service.test-namespace.svc",
		fmt.Sprintf("test-service.test-namespace.svc.%s", common.GetClusterSuffix()),
		"test-pod.test-service",
		"test-pod.test-service.test-namespace",
		"test-pod.test-service.test-namespace.svc",
		fmt.Sprintf("test-pod.test-service.test-namespace.svc.%s", common.GetClusterSuffix()),
	}
	if !reflect.DeepEqual(dnsNames, expectedDNSNames) {
		t.Error(dnsNames)
	}
}

func TestStatefulSetDNSNames(t *testing.T) {
	dnsNames := StatefulSetDNSNames("test-service", "test-namespace", int32(1))
	expectedDNSNames := []string{
		"test-service",
		"test-service.test-namespace",
		"test-service.test-namespace.svc",
		fmt.Sprintf("test-service.test-namespace.svc.%s", common.GetClusterSuffix()),
		"test-service-0.test-service",
		"test-service-0.test-service.test-namespace",
		"test-service-0.test-service.test-namespace.svc",
		fmt.Sprintf("test-service-0.test-service.test-namespace.svc.%s", common.GetClusterSuffix()),
	}
	if !reflect.DeepEqual(dnsNames, expectedDNSNames) {
		t.Error(dnsNames)
	}
}
