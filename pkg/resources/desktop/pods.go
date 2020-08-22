package desktop

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newDesktopPodForCR(cluster *v1alpha1.VDICluster, tmpl *v1alpha1.DesktopTemplate, instance *v1alpha1.Desktop) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetName(),
			Namespace:       instance.GetNamespace(),
			Labels:          cluster.GetDesktopLabels(instance),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: corev1.PodSpec{
			Hostname:           instance.GetName(),
			Subdomain:          instance.GetName(),
			ServiceAccountName: tmpl.GetDesktopServiceAccount(),
			SecurityContext:    tmpl.GetDesktopPodSecurityContext(),
			Volumes:            tmpl.GetDesktopVolumes(cluster, instance),
			ImagePullSecrets:   tmpl.GetDesktopPullSecrets(),
			Containers: []corev1.Container{
				tmpl.GetDesktopProxyContainer(),
				{
					Name:            "desktop",
					Image:           tmpl.GetDesktopImage(),
					ImagePullPolicy: tmpl.GetDesktopPullPolicy(),
					VolumeMounts:    tmpl.GetDesktopVolumeMounts(cluster, instance),
					SecurityContext: tmpl.GetDesktopContainerSecurityContext(),
					Env:             tmpl.GetDesktopEnvVars(instance),
					Lifecycle:       tmpl.GetLifecycle(),
					Resources:       tmpl.GetDesktopResources(),
				},
			},
		},
	}
}

func newServiceForCR(cluster *v1alpha1.VDICluster, instance *v1alpha1.Desktop) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetName(),
			Namespace:       instance.GetNamespace(),
			Labels:          cluster.GetDesktopLabels(instance),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: cluster.GetDesktopLabels(instance),
			Ports: []corev1.ServicePort{
				{
					Name:       "novnc-proxy",
					Port:       v1.WebPort,
					TargetPort: intstr.FromInt(v1.WebPort),
				},
			},
		},
	}
}
