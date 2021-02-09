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
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func newFakePVC() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-pvc",
			Namespace: "fake-namespace",
		},
		Spec: corev1.PersistentVolumeClaimSpec{},
	}
}

func TestReconcilePersistentVolumeClaim(t *testing.T) {
	c := getFakeClient(t)
	pvc := newFakePVC()
	if err := PersistentVolumeClaim(context.TODO(), testLogger, c, pvc); err != nil {
		t.Error("Expected no error, got:", err)
	}
	// Should be idempotent
	if err := PersistentVolumeClaim(context.TODO(), testLogger, c, pvc); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, pvc); err != nil {
		t.Error("Expected pvc to exist, got:", err)
	}
}
