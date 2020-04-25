package reconcile

import (
	"context"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var replicas int32 = 1

func newFakeDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-deployment",
			Namespace: "fake-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
	}
}

func TestReconcileDeployment(t *testing.T) {
	c := getFakeClient(t)
	deployment := newFakeDeployment()
	if err := ReconcileDeployment(testLogger, c, deployment, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	if err := c.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment); err != nil {
		t.Error("Expected deployment to exist, got:", err)
	}

	if err := ReconcileDeployment(testLogger, c, deployment, false); err != nil {
		t.Error("Expected no error, got", err)
	}

	deployment.Status = appsv1.DeploymentStatus{
		ReadyReplicas: 0,
	}
	c.Status().Update(context.TODO(), deployment)

	if err := ReconcileDeployment(testLogger, c, deployment, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	deployment.Status = appsv1.DeploymentStatus{
		ReadyReplicas: 1,
	}
	c.Status().Update(context.TODO(), deployment)
	if err := ReconcileDeployment(testLogger, c, deployment, true); err != nil {
		t.Error("Expected no error, got", err)
	}
}
