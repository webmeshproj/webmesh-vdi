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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFakeConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-configmap",
			Namespace: "fake-namespace",
		},
		Data: map[string]string{},
	}
}

func TestReconcileConfigMap(t *testing.T) {
	c := getFakeClient(t)
	cm := newFakeConfigMap()
	if err := ConfigMap(testLogger, c, cm); err != nil {
		t.Error("Expected no error, got:", err)
	}
	// should be idempotent
	cm = newFakeConfigMap()
	if err := ConfigMap(testLogger, c, cm); err != nil {
		t.Error("Expected no error, got:", err)
	}

	// another would trigger update (object metadata has changed)
	if err := ConfigMap(testLogger, c, cm); err != nil {
		t.Error("Expected no error, got:", err)
	}

}
