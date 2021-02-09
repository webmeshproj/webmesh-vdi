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
