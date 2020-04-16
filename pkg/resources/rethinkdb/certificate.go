package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDBCertForCR(instance *v1alpha1.VDICluster) *cm.Certificate {
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetRethinkDBName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("db"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.CertificateSpec{
			KeySize:    4096,
			CommonName: instance.GetRethinkDBName(),
			DNSNames: append(
				util.StatefulSetDNSNames(instance.GetRethinkDBName(), instance.GetCoreNamespace(), *instance.GetRethinkDBReplicas()),
				util.StatefulSetDNSNames(instance.GetRethinkDBProxyName(), instance.GetCoreNamespace(), *instance.GetRethinkDBProxyReplicas())...,
			),
			SecretName: instance.GetRethinkDBName(),
			Usages:     tlsutil.ServerMTLSUsages(),
			Subject: &cm.X509Subject{
				SerialNumber: "2",
			},
			IssuerRef: cmmeta.ObjectReference{
				Name: instance.GetCAName(),
				Kind: "ClusterIssuer",
			},
		},
	}
}

func newMgrClientCertForCR(instance *v1alpha1.VDICluster) *cm.Certificate {
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetRethinkDBClientName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("db"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.CertificateSpec{
			KeySize:    4096,
			CommonName: instance.GetRethinkDBClientName(),
			SecretName: instance.GetRethinkDBClientName(),
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
