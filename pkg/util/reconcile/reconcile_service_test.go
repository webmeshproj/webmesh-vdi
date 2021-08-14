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

	"github.com/kvdi/kvdi/pkg/util/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFakeService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-pod",
			Namespace: "fake-namespace",
		},
		Spec: corev1.ServiceSpec{},
	}
}

func TestReconcileService(t *testing.T) {
	c := getFakeClient(t)
	svc := newFakeService()
	if err := Service(context.TODO(), testLogger, c, svc); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := Service(context.TODO(), testLogger, c, newFakeService()); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := Service(context.TODO(), testLogger, c, svc); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}
}
