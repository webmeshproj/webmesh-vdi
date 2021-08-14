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

package reconcile

import (
	"testing"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"

	kappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	krbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// TODO: Apparently fake client will be deprecated and these tests should
// use envttest instead.

var testLogger = logf.Log.WithName("test")

func getFakeClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	rbacv1.AddToScheme(scheme)
	appv1.AddToScheme(scheme)
	desktopsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	kappsv1.AddToScheme(scheme)
	krbacv1.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}
