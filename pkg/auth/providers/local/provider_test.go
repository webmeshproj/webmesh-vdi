package local

import (
	"reflect"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TODO: Apparently fake client will be deprecated and these tests should
// use envttest instead.

func getFakeClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}

func TestNew(t *testing.T) {
	if reflect.TypeOf(New()) != reflect.TypeOf(&LocalAuthProvider{}) {
		t.Error("Someone messed with New")
	}
}

func TestSetup(t *testing.T) {
	cluster := &v1alpha1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Namespace = "test-namespace"
	provider := New()
	if err := provider.Setup(getFakeClient(t), cluster); err != nil {
		t.Error("No error should happen when setting up the local auth provider")
	}
}
