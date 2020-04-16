package app

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newAppDeploymentForCR(instance *v1alpha1.VDICluster) *appsv1.Deployment {
	args := []string{"--vdi-cluster", instance.GetName()}
	if instance.EnableCORS() {
		args = append(args, "--enable-cors")
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instance.GetAppName(),
			Namespace:       instance.GetCoreNamespace(),
			Labels:          instance.GetComponentLabels("app"),
			Annotations:     instance.GetAnnotations(),
			OwnerReferences: instance.OwnerReferences(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.GetAppReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: instance.GetComponentLabels("app"),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: instance.GetComponentLabels("app"),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.GetAppName(),
					SecurityContext:    instance.GetAppSecurityContext(),
					Volumes: []corev1.Volume{
						{
							Name: "tls-server",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: instance.GetAppName(),
								},
							},
						},
						{
							Name: "tls-client",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-client", instance.GetAppName()),
								},
							},
						},
					},
					ImagePullSecrets: instance.GetPullSecrets(),
					Containers: []corev1.Container{
						{
							Name:            "app",
							Image:           instance.GetAppImage(),
							ImagePullPolicy: instance.GetAppPullPolicy(),
							Resources:       instance.GetAppResources(),
							Args:            args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: v1alpha1.WebPort,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tls-server",
									MountPath: v1alpha1.ServerCertificateMountPath,
									ReadOnly:  true,
								},
								{
									Name:      "tls-client",
									MountPath: v1alpha1.ClientCertificateMountPath,
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	}
}
