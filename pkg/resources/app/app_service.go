package app

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newAppServiceForCR(instance *v1alpha1.VDICluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: instance.GetComponentLabels("app"),
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       v1.PublicWebPort,
					TargetPort: intstr.FromInt(v1.WebPort),
				},
			},
		},
	}
}
