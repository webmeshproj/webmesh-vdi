package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// rethinkDBStartScript is the start script used on rethinkdb cluster nodes
var rethinkDBStartScript = `
set -exo pipefail

ORDINAL=$(echo "${POD_NAME}" | rev | cut -d "-" -f1 | rev)

ARGS="
--bind all \
--bind-http 127.0.0.1 \
--directory /data/rethinkdb_data \
--server-name ${POD_NAME} \
--server-tag ${POD_NAME} \
--server-tag ${NODE_NAME} \
--canonical-address ${POD_IP} \
--http-tls-key ${TLS_KEY} \
--http-tls-cert ${TLS_CERT} \
--driver-tls-key ${TLS_KEY} \
--driver-tls-cert ${TLS_CERT} \
--driver-tls-ca ${TLS_CA_CERT} \
--cluster-tls-key ${TLS_KEY} \
--cluster-tls-cert ${TLS_CERT} \
--cluster-tls-ca ${TLS_CA_CERT}
"

if [[ "${ORDINAL}" == "0" ]]; then
	echo "Start single/master instance"
	exec rethinkdb ${ARGS}
else
	while ! getent hosts ${SERVICE_NAME}.${POD_NAMESPACE} ; do sleep 3 ; done
	ENDPOINT="${SERVICE_NAME}-0.${SERVICE_NAME}.${POD_NAMESPACE}.svc.${CLUSTER_SUFFIX}:29015"
	echo "Join to ${SERVICE_NAME} on ${ENDPOINT}"
	exec rethinkdb --join ${ENDPOINT} ${ARGS}
fi
`

// newStatefulSetForCR returns a new rethinkdb statefulset configuration for the given
// AndroidFarm instance.
func newStatefulSetForCR(cr *v1alpha1.VDICluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.GetRethinkDBName(),
			Namespace:       cr.GetCoreNamespace(),
			Labels:          cr.GetComponentLabels("rethinkdb"),
			OwnerReferences: cr.OwnerReferences(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: cr.GetRethinkDBReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: cr.GetComponentLabels("rethinkdb"),
			},
			ServiceName:          cr.GetRethinkDBName(),
			VolumeClaimTemplates: cr.GetRethinkDBVolumeClaims(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cr.GetComponentLabels("rethinkdb"),
				},
				Spec: corev1.PodSpec{
					SecurityContext: cr.GetAppSecurityContext(),
					Volumes:         cr.GetRethinkDBVolumes(),
					Containers: []corev1.Container{
						{
							Name:            "rethinkdb",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Image:           cr.GetRethinkDBImage(),
							Env: append(cr.RethinkDBEnvVars(), corev1.EnvVar{
								Name:  "CLUSTER_SUFFIX",
								Value: common.GetClusterSuffix(),
							}),
							VolumeMounts: cr.GetRethinkDBVolumeMounts(),
							Command:      []string{"/bin/bash", "-c"},
							Args:         []string{rethinkDBStartScript},
							Ports: []corev1.ContainerPort{
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
								InitialDelaySeconds: 5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
								PeriodSeconds:       10,
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.Parse("cluster-port"),
									},
								},
							},
							Resources: cr.GetRethinkDBResources(),
						},
					},
				},
			},
		},
	}
}
