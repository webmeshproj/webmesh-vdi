package reconcile

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"

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
	if err := Service(testLogger, c, svc); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := Service(testLogger, c, newFakeService()); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := Service(testLogger, c, svc); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}
}
