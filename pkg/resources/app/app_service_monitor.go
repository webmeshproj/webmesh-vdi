package app

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newAppServiceMonitorForCR(instance *v1alpha1.VDICluster) *promv1.ServiceMonitor {
	return &promv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetServiceMonitorLabels(),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: promv1.ServiceMonitorSpec{
			NamespaceSelector: promv1.NamespaceSelector{
				MatchNames: []string{instance.GetCoreNamespace()},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					v1.VDIClusterLabel: instance.GetName(),
					v1.ComponentLabel:  "app",
				},
			},
			Endpoints: []promv1.Endpoint{
				{
					Port:     "web",
					Path:     "/api/metrics",
					Interval: "10s",
					Scheme:   "https",
					TLSConfig: &promv1.TLSConfig{
						InsecureSkipVerify: true,
					},
				},
			},
		},
	}
}
