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

func newFakeCertificate() *cm.Certificate {
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-certificate",
			Namespace: "fake-namespace",
		},
		Spec: cm.CertificateSpec{
			KeySize:    4096,
			CommonName: "fake-common-name",
			SecretName: "fake-secret",
			IssuerRef: cmmeta.ObjectReference{
				Name: "fake-issuer",
				Kind: "ClusterIssuer",
			},
		},
	}
}

func TestReconcileCertificate(t *testing.T) {
	fakeClient := getFakeClient(t)
	cert := newFakeCertificate()
	// Should requeue immediately
	if err := ReconcileCertificate(testLogger, fakeClient, cert, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}
	// should go through fully without wait
	if err := ReconcileCertificate(testLogger, fakeClient, cert, false); err != nil {
		t.Error("Expected no error, got:", err)
	}
	// Should wait because status isn't ready
	if err := ReconcileCertificate(testLogger, fakeClient, cert, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	fakeClient.Get(context.TODO(), types.NamespacedName{Name: cert.Name, Namespace: cert.Namespace}, cert)
	cert.Status = cm.CertificateStatus{
		Conditions: []cm.CertificateCondition{
			{Type: cm.CertificateConditionReady, Status: cmmeta.ConditionFalse},
		},
	}
	fakeClient.Status().Update(context.TODO(), cert)

	// Should still wait because status isn't ready
	if err := ReconcileCertificate(testLogger, fakeClient, cert, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	fakeClient.Get(context.TODO(), types.NamespacedName{Name: cert.Name, Namespace: cert.Namespace}, cert)
	cert.Status = cm.CertificateStatus{
		Conditions: []cm.CertificateCondition{
			{Type: cm.CertificateConditionReady, Status: cmmeta.ConditionTrue},
		},
	}
	fakeClient.Status().Update(context.TODO(), cert)

	// Should no longer wait
	if err := ReconcileCertificate(testLogger, fakeClient, cert, true); err != nil {
		t.Error("Expected no error, got", err)
	}
}
