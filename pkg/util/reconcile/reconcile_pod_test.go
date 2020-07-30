package reconcile

import (
	"context"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"

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

func TestReconcilePod(t *testing.T) {
	c := getFakeClient(t)
	pod := newFakePod()

	if created, err := Pod(testLogger, c, pod); err != nil {
		t.Error("Expected no error, got:", err)
	} else if !created {
		t.Error("Expected created to be true")
	}

	if created, err := Pod(testLogger, c, newFakePod()); err != nil {
		t.Error("Expected no error, got:", err)
	} else if created {
		t.Error("Expected created to be false")
	}

	now := metav1.Now()
	pod.SetDeletionTimestamp(&now)
	c.Update(context.TODO(), pod)

	if _, err := Pod(testLogger, c, newFakePod()); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	pod.SetDeletionTimestamp(nil)
	c.Update(context.TODO(), pod)

	// expect delete and requeue for changed pod
	if _, err := Pod(testLogger, c, pod); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}
}
