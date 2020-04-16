package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// newServiceForCR returns a new headless service for the rehtinkdb cluster
func newServiceForCR(cr *v1alpha1.VDICluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.GetRethinkDBName(),
			Namespace:       cr.GetCoreNamespace(),
			Labels:          cr.GetComponentLabels("rethinkdb"),
			OwnerReferences: cr.OwnerReferences(),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "driver-port",
					Port:       int32(v1alpha1.RethinkDBDriverPort),
					TargetPort: intstr.FromInt(v1alpha1.RethinkDBDriverPort),
				},
				{
					Name:       "cluster-port",
					Port:       int32(v1alpha1.RethinkDBClusterPort),
					TargetPort: intstr.FromInt(v1alpha1.RethinkDBClusterPort),
				},
			},
			Selector: cr.GetComponentLabels("rethinkdb"),
		},
	}
}

// newProxyServiceForCR returns a headless service for the rehtinkdb proxies
func newProxyServiceForCR(cr *v1alpha1.VDICluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.GetRethinkDBProxyName(),
			Namespace:       cr.GetCoreNamespace(),
			Labels:          cr.GetComponentLabels("rethinkdb-proxy"),
			OwnerReferences: cr.OwnerReferences(),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "admin-port",
					Port:       int32(v1alpha1.RethinkDBAdminPort),
					TargetPort: intstr.FromInt(v1alpha1.RethinkDBAdminPort),
				},
				{
					Name:       "driver-port",
					Port:       int32(v1alpha1.RethinkDBDriverPort),
					TargetPort: intstr.FromInt(v1alpha1.RethinkDBDriverPort),
				},
				{
					Name:       "cluster-port",
					Port:       int32(v1alpha1.RethinkDBClusterPort),
					TargetPort: intstr.FromInt(v1alpha1.RethinkDBClusterPort),
				},
			},
			Selector: cr.GetComponentLabels("rethinkdb-proxy"),
		},
	}
}
