package reconcile

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// TODO: Apparently fake client will be deprecated and these tests should
// use envttest instead.

var testLogger = logf.Log.WithName("test")

func getFakeClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}
