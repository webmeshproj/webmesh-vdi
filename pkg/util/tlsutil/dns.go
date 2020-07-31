package tlsutil

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

// DNSNames returns the cluster DNS names for the given service name and namespace.
func DNSNames(svcName, svcNamespace string) []string {
	return []string{
		svcName,
		fmt.Sprintf("%s.%s", svcName, svcNamespace),
		fmt.Sprintf("%s.%s.svc", svcName, svcNamespace),
		fmt.Sprintf("%s.%s.svc.%s", svcName, svcNamespace, common.GetClusterSuffix()),
	}
}

// HeadlessDNSNames returns the DNS names for a pod in the given headless service name
// and namespace.
func HeadlessDNSNames(podName, svcName, svcNamespace string) []string {
	return append(DNSNames(svcName, svcNamespace), []string{
		fmt.Sprintf("%s.%s", podName, svcName),
		fmt.Sprintf("%s.%s.%s", podName, svcName, svcNamespace),
		fmt.Sprintf("%s.%s.%s.svc", podName, svcName, svcNamespace),
		fmt.Sprintf("%s.%s.%s.svc.%s", podName, svcName, svcNamespace, common.GetClusterSuffix()),
	}...)
}
