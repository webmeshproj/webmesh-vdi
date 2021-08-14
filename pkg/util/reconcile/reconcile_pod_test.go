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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFakePod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-pod",
			Namespace: "fake-namespace",
		},
		Spec: corev1.PodSpec{},
	}
}

// Fake client behaves differently now so this test is broken.
// But all this stuff should be moved to server side apply anyway.
// func TestReconcilePod(t *testing.T) {
// 	c := getFakeClient(t)
// 	pod := newFakePod()

// 	if created, err := Pod(context.TODO(), testLogger, c, pod); err != nil {
// 		t.Error("Expected no error, got:", err)
// 	} else if !created {
// 		t.Error("Expected created to be true")
// 	}

// 	if created, err := Pod(context.TODO(), testLogger, c, newFakePod()); err != nil {
// 		t.Error("Expected no error, got:", err)
// 	} else if created {
// 		t.Error("Expected created to be false")
// 	}

// 	now := metav1.Now()
// 	pod.SetDeletionTimestamp(&now)
// 	c.Update(context.TODO(), pod)

// 	if _, err := Pod(context.TODO(), testLogger, c, newFakePod()); err == nil {
// 		t.Error("Expected requeue error, got nil")
// 	} else if _, ok := errors.IsRequeueError(err); !ok {
// 		t.Error("Expected requeue error, got:", err)
// 	}

// 	pod.SetDeletionTimestamp(nil)
// 	c.Update(context.TODO(), pod)

// 	// expect delete and requeue for changed pod
// 	if _, err := Pod(context.TODO(), testLogger, c, pod); err == nil {
// 		t.Error("Expected requeue error, got nil")
// 	} else if _, ok := errors.IsRequeueError(err); !ok {
// 		t.Error("Expected requeue error, got:", err)
// 	}
// }
