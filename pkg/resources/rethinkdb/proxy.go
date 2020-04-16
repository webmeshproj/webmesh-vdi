package rethinkdb

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// proxyStartScript is the start script used for the proxy instances
var proxyStartScript = `
set -exo pipefail
while ! getent hosts ${WAIT_SERVERS} ; do sleep 3 ; done
exec rethinkdb proxy \
  ${JOIN_SERVERS} \
  --bind all \
  --canonical-address ${POD_IP} \
  --log-file /data/rethinkdb_data/log_file \
  --http-tls-key ${TLS_KEY} \
  --http-tls-cert ${TLS_CERT} \
  --driver-tls-key ${TLS_KEY} \
  --driver-tls-cert ${TLS_CERT} \
  --driver-tls-ca ${TLS_CA_CERT} \
  --cluster-tls-key ${TLS_KEY} \
  --cluster-tls-cert ${TLS_CERT} \
  --cluster-tls-ca ${TLS_CA_CERT}
`

func (r *RethinkDBReconciler) reconcileProxy(reqLogger logr.Logger, instance *v1alpha1.VDICluster) error {
	clusterSuffix := util.GetClusterSuffix()
	name := instance.GetRethinkDBName()
	namespace := instance.GetCoreNamespace()
	joinStr := ""
	waitStr := ""
	for i := int32(0); i < *instance.GetRethinkDBReplicas(); i++ {
		addr := fmt.Sprintf("%s-%d.%s.%s.svc.%s", name, i, name, namespace, clusterSuffix)
		joinStr = joinStr + fmt.Sprintf(" --join %s:29015 ", addr)
		waitStr = waitStr + fmt.Sprintf(" %s ", addr)
	}
	ss := newProxyStatefulSetForCR(instance, joinStr, waitStr)
	if err := reconcile.ReconcileStatefulSet(reqLogger, r.client, ss, true); err != nil {
		return err
	}
	return nil
}

// newProxyStatefulSetForCR returns a new proxy statefulset for an AndroidFarm instance.
func newProxyStatefulSetForCR(cr *v1alpha1.VDICluster, joinStr, waitStr string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.GetRethinkDBProxyName(),
			Namespace:       cr.GetCoreNamespace(),
			Labels:          cr.GetComponentLabels("rethinkdb-proxy"),
			OwnerReferences: cr.OwnerReferences(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: cr.GetRethinkDBProxyReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: cr.GetComponentLabels("rethinkdb-proxy"),
			},
			ServiceName: cr.GetRethinkDBProxyName(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cr.GetComponentLabels("rethinkdb-proxy"),
				},
				Spec: corev1.PodSpec{
					SecurityContext: cr.GetAppSecurityContext(),
					Volumes: []corev1.Volume{
						{
							Name:         "rethinkdb-proxy-log",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
						{
							Name: "rethinkdb-proxy-tls",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cr.GetRethinkDBName(),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "rethinkdb-proxy",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Image:           cr.GetRethinkDBImage(),
							Command:         []string{"/bin/bash", "-c"},
							Args:            []string{proxyStartScript},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "rethinkdb-proxy-log",
									MountPath: "/data/rethinkdb_data",
								},
								{
									Name:      "rethinkdb-proxy-tls",
									MountPath: v1alpha1.ServerCertificateMountPath,
								},
							},
							Env: append(cr.RethinkDBEnvVars(), []corev1.EnvVar{
								{
									Name:  "JOIN_SERVERS",
									Value: joinStr,
								},
								{
									Name:  "WAIT_SERVERS",
									Value: waitStr,
								},
								{
									Name:  "CLUSTER_SUFFIX",
									Value: util.GetClusterSuffix(),
								},
							}...),
							Ports: []corev1.ContainerPort{
								{
									Name:          "admin-port",
									ContainerPort: v1alpha1.RethinkDBAdminPort,
								},
								{
									Name:          "driver-port",
									ContainerPort: v1alpha1.RethinkDBDriverPort,
								},
								{
									Name:          "cluster-port",
									ContainerPort: v1alpha1.RethinkDBClusterPort,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Scheme: "HTTPS",
										Port:   intstr.Parse("admin-port"),
									},
								},
							},
							Resources: cr.GetRethinkDBProxyResources(),
						},
					},
				},
			},
		},
	}
}
