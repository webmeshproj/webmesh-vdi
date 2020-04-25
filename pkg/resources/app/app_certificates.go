package app

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Allow custom frontend TLS
func newAppCertForCR(instance *v1alpha1.VDICluster) *cm.Certificate {
	dnsNames := tlsutil.DNSNames(instance.GetAppName(), instance.GetCoreNamespace())
	if instance.GetAppExternalHostname() != "" {
		dnsNames = append(dnsNames, instance.GetAppExternalHostname())
	}
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.CertificateSpec{
			KeySize:    4096,
			CommonName: instance.GetAppName(),
			DNSNames:   dnsNames,
			SecretName: instance.GetAppName(),
			IssuerRef: cmmeta.ObjectReference{
				Name: instance.GetCAName(),
				Kind: "ClusterIssuer",
			},
		},
	}
}

func newAppClientCertForCR(instance *v1alpha1.VDICluster) *cm.Certificate {
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-client", instance.GetAppName()),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.CertificateSpec{
			KeySize:    4096,
			CommonName: instance.GetAppName(),
			DNSNames:   tlsutil.DNSNames(instance.GetAppName(), instance.GetCoreNamespace()),
			SecretName: fmt.Sprintf("%s-client", instance.GetAppName()),
			Usages:     tlsutil.ClientMTLSUsages(),
			Subject: &cm.X509Subject{
				SerialNumber: "3",
			},
			IssuerRef: cmmeta.ObjectReference{
				Name: instance.GetCAName(),
				Kind: "ClusterIssuer",
			},
		},
	}
}
