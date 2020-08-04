package app

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GrafanaDashboard is the JSON of the Grafana dashboard. This file is giant
// so as not to muddy up the code it is set via build flags.
var GrafanaDashboard string

// GrafanaDatasourceTmpl defines the prometheus datasource configuration to
// provide to the grafana image.
var GrafanaDatasourceTmpl = `apiVersion: 1
datasources:
  - name: prometheus
    type: prometheus
    access: proxy
    url: %s.%s.svc:9090
`

func newGrafanaConfigForCR(instance *v1alpha1.VDICluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-grafana", instance.GetAppName()),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("metrics"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Data: map[string]string{
			"dashboard.json":  GrafanaDashboard,
			"datasource.yaml": fmt.Sprintf(GrafanaDatasourceTmpl, instance.GetPrometheusName(), instance.GetCoreNamespace()),
		},
	}
}

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
					v1.ComponentLabel:  "app",
					v1.VDIClusterLabel: instance.GetName(),
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

func newPrometheusForCR(instance *v1alpha1.VDICluster) *promv1.Prometheus {
	return &promv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetPrometheusName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("metrics"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: promv1.PrometheusSpec{
			ServiceMonitorSelector: &metav1.LabelSelector{
				MatchLabels: instance.GetServiceMonitorLabels(),
			},
			ServiceAccountName: instance.GetAppName(),
			Resources:          instance.GetPrometheusResources(),
		},
	}
}

func newPrometheusServiceForCR(instance *v1alpha1.VDICluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetPrometheusName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("metrics"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"prometheus": instance.GetPrometheusName(),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
				},
			},
		},
	}
}
