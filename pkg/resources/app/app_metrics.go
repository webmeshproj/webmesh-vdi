/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package app

import (
	_ "embed"
	"fmt"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GrafanaDashboard is the JSON of the Grafana dashboard.
//go:embed grafana-dashboard.json
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

func newGrafanaConfigForCR(instance *appv1.VDICluster) *corev1.ConfigMap {
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

func newAppServiceMonitorForCR(instance *appv1.VDICluster) *promv1.ServiceMonitor {
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
						SafeTLSConfig: promv1.SafeTLSConfig{
							InsecureSkipVerify: true,
						},
					},
				},
			},
		},
	}
}

func newPrometheusForCR(instance *appv1.VDICluster) *promv1.Prometheus {
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

func newPrometheusServiceForCR(instance *appv1.VDICluster) *corev1.Service {
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
