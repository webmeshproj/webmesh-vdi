/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

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
