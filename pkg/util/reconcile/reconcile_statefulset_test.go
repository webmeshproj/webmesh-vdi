package reconcile

import (
	"context"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var ssReplicas int32 = 1

func newFakeStatefulSet() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-statefulset",
			Namespace: "fake-namespace",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &ssReplicas,
		},
	}
}

func TestReconcileStatefulSet(t *testing.T) {
	c := getFakeClient(t)
	ss := newFakeStatefulSet()

	if err := ReconcileStatefulSet(testLogger, c, ss, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	if err := c.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, ss); err != nil {
		t.Error("Expected ss to exist, got:", err)
	}

	if err := ReconcileStatefulSet(testLogger, c, ss, false); err != nil {
		t.Error("Expected no error, got", err)
	}

	ss.Status = appsv1.StatefulSetStatus{
		ReadyReplicas: 0,
	}
	c.Status().Update(context.TODO(), ss)

	if err := ReconcileStatefulSet(testLogger, c, ss, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	ss.Status = appsv1.StatefulSetStatus{
		ReadyReplicas: 1,
	}
	c.Status().Update(context.TODO(), ss)
	if err := ReconcileStatefulSet(testLogger, c, ss, true); err != nil {
		t.Error("Expected no error, got", err)
	}
}
