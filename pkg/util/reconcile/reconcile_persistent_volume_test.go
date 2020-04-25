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
	if err := ReconcilePersistentVolumeClaim(testLogger, c, pvc); err != nil {
		t.Error("Expected no error, got:", err)
	}
	// Should be idempotent
	if err := ReconcilePersistentVolumeClaim(testLogger, c, pvc); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, pvc); err != nil {
		t.Error("Expected pvc to exist, got:", err)
	}
}
