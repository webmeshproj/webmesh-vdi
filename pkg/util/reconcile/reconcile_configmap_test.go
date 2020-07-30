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
