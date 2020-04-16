package desktop

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDesktopProxyCert(cluster *v1alpha1.VDICluster, instance *v1alpha1.Desktop) *cm.Certificate {
	return &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetName(),
			Namespace:       instance.GetNamespace(),
			Labels:          cluster.GetDesktopLabels(instance),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: cm.CertificateSpec{
			KeySize:    4096,
			CommonName: "desktop",
			DNSNames:   util.HeadlessDNSNames(instance.GetName(), instance.GetName(), instance.GetNamespace()),
			SecretName: instance.GetName(),
			Usages:     tlsutil.ServerMTLSUsages(),
			IssuerRef: cmmeta.ObjectReference{
				Name: cluster.GetCAName(),
				Kind: "ClusterIssuer",
			},
		},
	}
}
