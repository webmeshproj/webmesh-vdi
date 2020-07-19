package k8sutil

import (
	"context"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getFakeClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}

func TestLookupClusterByName(t *testing.T) {
	c := getFakeClient(t)
	c.Create(context.TODO(), &v1alpha1.VDICluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-cluster",
		},
	})
	if _, err := LookupClusterByName(c, "fake-cluster"); err != nil {
		t.Error("Expected no error for existing cluster, got:", err)
	}
	if _, err := LookupClusterByName(c, "another-cluster"); err == nil {
		t.Error("Expected error for non-existing cluster, got nil")
	}
}

func TestIsMarkedForDeletion(t *testing.T) {
	cr := &v1alpha1.VDICluster{}
	now := metav1.Now()
	cr.SetDeletionTimestamp(&now)
	if !IsMarkedForDeletion(cr) {
		t.Error("Expected CR to be marked for deletion")
	}
}

func TestCreationSpecAnnotations(t *testing.T) {
	cr := &v1alpha1.VDICluster{}
	cr.Name = "test-cluster"
	cr.Namespace = "test-namespace"
	if err := SetCreationSpecAnnotation(&cr.ObjectMeta, cr); err != nil {
		t.Error("Expected no error, got:", err)
	}
	if cr.GetAnnotations() == nil {
		t.Error("Expected a new set of annotations, got nil")
	}
	if _, ok := cr.Annotations[v1.CreationSpecAnnotation]; !ok {
		t.Error("Expected creation spec annotation to be set")
	}

	newCR := &v1alpha1.VDICluster{}
	newCR.Name = "test-cluster"
	newCR.Namespace = "test-namespace"

	if CreationSpecsEqual(cr.ObjectMeta, newCR.ObjectMeta) {
		t.Error("Expected non equal since no annotations set")
	}
	if CreationSpecsEqual(newCR.ObjectMeta, cr.ObjectMeta) {
		t.Error("Expected non equal since no annotations set")
	}

	if err := SetCreationSpecAnnotation(&newCR.ObjectMeta, newCR); err != nil {
		t.Error("Expected no error, got:", err)
	}

	if !CreationSpecsEqual(newCR.ObjectMeta, cr.ObjectMeta) {
		t.Error("Expected equal creation specs")
	}
}
