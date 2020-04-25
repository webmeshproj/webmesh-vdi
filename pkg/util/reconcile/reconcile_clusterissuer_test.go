package reconcile

import (
	"context"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func newFakeIssuer() *cm.ClusterIssuer {
	return &cm.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-issuer",
			Namespace: "fake-namespace",
		},
		Spec: cm.IssuerSpec{},
	}
}

func TestReconcileClusterIssuer(t *testing.T) {
	fakeClient := getFakeClient(t)
	issuer := newFakeIssuer()
	// Should requeue immediately
	if err := ReconcileClusterIssuer(testLogger, fakeClient, issuer, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}
	// should go through fully without wait
	if err := ReconcileClusterIssuer(testLogger, fakeClient, issuer, false); err != nil {
		t.Error("Expected no error, got:", err)
	}
	// Should wait because status isn't ready
	if err := ReconcileClusterIssuer(testLogger, fakeClient, issuer, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	fakeClient.Get(context.TODO(), types.NamespacedName{Name: issuer.Name, Namespace: issuer.Namespace}, issuer)
	issuer.Status = cm.IssuerStatus{
		Conditions: []cm.IssuerCondition{
			{Type: cm.IssuerConditionReady, Status: cmmeta.ConditionFalse},
		},
	}
	fakeClient.Status().Update(context.TODO(), issuer)

	// Should still wait because status isn't ready
	if err := ReconcileClusterIssuer(testLogger, fakeClient, issuer, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	fakeClient.Get(context.TODO(), types.NamespacedName{Name: issuer.Name, Namespace: issuer.Namespace}, issuer)
	issuer.Status = cm.IssuerStatus{
		Conditions: []cm.IssuerCondition{
			{Type: cm.IssuerConditionReady, Status: cmmeta.ConditionTrue},
		},
	}
	fakeClient.Status().Update(context.TODO(), issuer)

	// Should no longer wait
	if err := ReconcileClusterIssuer(testLogger, fakeClient, issuer, true); err != nil {
		t.Error("Expected no error, got", err)
	}
}
