package tlsutil

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

func DNSNames(svcName, svcNamespace string) []string {
	return []string{
		svcName,
		fmt.Sprintf("%s.%s", svcName, svcNamespace),
		fmt.Sprintf("%s.%s.svc", svcName, svcNamespace),
		fmt.Sprintf("%s.%s.svc.%s", svcName, svcNamespace, common.GetClusterSuffix()),
	}
}

func HeadlessDNSNames(podName, svcName, svcNamespace string) []string {
	return append(DNSNames(svcName, svcNamespace), []string{
		fmt.Sprintf("%s.%s", podName, svcName),
		fmt.Sprintf("%s.%s.%s", podName, svcName, svcNamespace),
		fmt.Sprintf("%s.%s.%s.svc", podName, svcName, svcNamespace),
		fmt.Sprintf("%s.%s.%s.svc.%s", podName, svcName, svcNamespace, common.GetClusterSuffix()),
	}...)
}

func StatefulSetDNSNames(svcName, svcNamespace string, replicas int32) []string {
	dnsNames := DNSNames(svcName, svcNamespace)
	for i := int32(0); i < replicas; i++ {
		podName := fmt.Sprintf("%s-%d", svcName, i)
		dnsNames = append(dnsNames,
			fmt.Sprintf("%s.%s", podName, svcName),
			fmt.Sprintf("%s.%s.%s", podName, svcName, svcNamespace),
			fmt.Sprintf("%s.%s.%s.svc", podName, svcName, svcNamespace),
			fmt.Sprintf("%s.%s.%s.svc.%s", podName, svcName, svcNamespace, common.GetClusterSuffix()),
		)
	}
	return dnsNames
}
