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
	"context"
	"strings"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var testLogger = logf.Log.WithName("test")

func newReconciler(t *testing.T) *Reconciler {
	t.Helper()
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	promv1.AddToScheme(scheme)
	return New(fake.NewFakeClientWithScheme(scheme), scheme)
}

func newCluster(t *testing.T) *v1alpha1.VDICluster {
	t.Helper()
	cluster := &v1alpha1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Spec = v1alpha1.VDIClusterSpec{
		Metrics: &v1alpha1.MetricsConfig{
			ServiceMonitor: &v1alpha1.ServiceMonitorConfig{Create: true},
			Prometheus:     &v1alpha1.PrometheusConfig{Create: true},
			Grafana:        &v1alpha1.GrafanaConfig{Enabled: true},
		},
	}
	return cluster
}

// TestReconcile tests the reconcile workflow.
// TODO: need to add ability to wrap hardcoded errors in a requeue error
// for better testing.
func TestReconcile(t *testing.T) {
	r := newReconciler(t)
	cluster := newCluster(t)

	// expect everything to be created except for a deployment requeue
	if err := r.Reconcile(testLogger, cluster); err == nil {
		t.Fatal("Expected error got nil")
	} else if qerr, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	} else if !strings.Contains(qerr.Error(), "deployment with wait") {
		t.Error("Expected error from waiting for deployment, got:", err)
	}

	// should keep happening until the deployment is ready
	if err := r.Reconcile(testLogger, cluster); err == nil {
		t.Fatal("Expected error got nil")
	} else if qerr, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	} else if !strings.Contains(qerr.Error(), "to be ready") {
		t.Error("Expected error from waiting for deployment, got:", err)
	}

	// update the deployment status
	deployment := &appsv1.Deployment{}
	nn := types.NamespacedName{Name: cluster.GetAppName(), Namespace: cluster.GetCoreNamespace()}
	if err := r.client.Get(context.TODO(), nn, deployment); err != nil {
		t.Fatal(err)
	}
	deployment.Status = appsv1.DeploymentStatus{ReadyReplicas: 1}
	if err := r.client.Status().Update(context.TODO(), deployment); err != nil {
		t.Fatal(err)
	}

	// should finish
	// TODO: check created resources
	if err := r.Reconcile(testLogger, cluster); err != nil {
		t.Error("Expected reconcile to complete successfully")
	}
}
