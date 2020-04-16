package pki

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newSignerForCR(instance *v1alpha1.VDICluster) *cm.ClusterIssuer {
	return &cm.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetSignerName(),
			Namespace:       metav1.NamespaceAll,
			Labels:          instance.GetComponentLabels("pki"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.IssuerSpec{
			IssuerConfig: cm.IssuerConfig{
				SelfSigned: &cm.SelfSignedIssuer{},
			},
		},
	}
}

func newCAForCR(instance *v1alpha1.VDICluster) *cm.Certificate {
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetCAName(),
			Namespace:       instance.GetCertManagerNamespace(),
			Labels:          instance.GetComponentLabels("pki"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.CertificateSpec{
			KeySize: 4096,
			IsCA:    true,
			Subject: &cm.X509Subject{
				SerialNumber: "1",
			},
			Usages:     tlsutil.CAUsages(),
			CommonName: instance.GetCAName(),
			DNSNames:   []string{instance.GetCAName()},
			SecretName: instance.GetCAName(),
			IssuerRef: cmmeta.ObjectReference{
				Name: instance.GetSignerName(),
				Kind: "ClusterIssuer",
			},
		},
	}
}

func newIssuerForCR(instance *v1alpha1.VDICluster) *cm.ClusterIssuer {
	return &cm.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetCAName(),
			Namespace:       metav1.NamespaceAll,
			Labels:          instance.GetComponentLabels("pki"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.IssuerSpec{
			IssuerConfig: cm.IssuerConfig{
				CA: &cm.CAIssuer{
					SecretName: instance.GetCAName(),
				},
			},
		},
	}
}
