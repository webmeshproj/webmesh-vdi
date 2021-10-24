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

package desktop

import (
	"context"
	"strings"
	"testing"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"

	"github.com/kvdi/kvdi/pkg/util/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var testLogger = logf.Log.WithName("test")

func newReconciler(t *testing.T) *Reconciler {
	t.Helper()
	scheme := runtime.NewScheme()
	appv1.AddToScheme(scheme)
	desktopsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	return New(fake.NewClientBuilder().WithScheme(scheme).Build(), scheme)
}

func newCluster(t *testing.T) *appv1.VDICluster {
	t.Helper()
	cluster := &appv1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Spec = appv1.VDIClusterSpec{
		UserdataSpec: &appv1.UserdataSpec{
			PersistentVolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{"storage": resource.MustParse("10Gi")},
				},
			}},
	}
	return cluster
}

func newDesktop(t *testing.T) *desktopsv1.Session {
	t.Helper()
	desktop := &desktopsv1.Session{}
	desktop.Name = "test-desktop"
	desktop.Namespace = "test-namespace"
	desktop.Spec = desktopsv1.SessionSpec{
		Template:   "test-template",
		VDICluster: "test-cluster",
	}
	return desktop
}

func newTemplate(t *testing.T) *desktopsv1.Template {
	t.Helper()
	tmpl := &desktopsv1.Template{}
	tmpl.Name = "test-template"
	return tmpl
}

// TestReconcile tests the reconcile workflow.
// TODO: need to add ability to wrap hardcoded errors in a requeue error
// for better testing.
func TestReconcile(t *testing.T) {
	r := newReconciler(t)
	desktop := newDesktop(t)
	cluster := newCluster(t)
	if err := r.client.Create(context.TODO(), desktop); err != nil {
		t.Fatal(err)
	}

	// Test missing dependent resources
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if client.IgnoreNotFound(err) != nil {
			t.Fatal(err)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// create the desktop template
	r.client.Create(context.TODO(), newTemplate(t))

	// cluster should now be not found
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if client.IgnoreNotFound(err) != nil {
			t.Fatal(err)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// create the cluster
	r.client.Create(context.TODO(), cluster)

	// now test actual reconciliation

	// Error should be waiting for service ip
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if qerr, ok := errors.IsRequeueError(err); !ok {
			t.Error("Expected requeue error, got:", err)
		} else if !strings.Contains(qerr.Error(), "assigned an IP") {
			t.Error("Expected waiting for service IP, got:", qerr)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// assign the svc an IP
	svc := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: desktop.GetName(), Namespace: desktop.GetNamespace()}, svc); err != nil {
		t.Fatal(err)
	}
	svc.Spec.ClusterIP = "127.0.0.1"
	if err := r.client.Update(context.TODO(), svc); err != nil {
		t.Fatal(err)
	}

	// error should be waiting for pod to be in running phase
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if qerr, ok := errors.IsRequeueError(err); !ok {
			t.Error("Expected requeue error, got:", err)
		} else if !strings.Contains(qerr.Error(), "not in running phase") {
			t.Error("Expected waiting for desktop running, got:", qerr)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// set the pod phase to running
	pod := &corev1.Pod{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: desktop.GetName(), Namespace: desktop.GetNamespace()}, pod); err != nil {
		t.Fatal(err)
	}
	pod.Status = corev1.PodStatus{
		Phase: corev1.PodRunning,
		ContainerStatuses: []corev1.ContainerStatus{
			{
				State: corev1.ContainerState{Running: nil},
			},
		},
	}
	if err := r.client.Status().Update(context.TODO(), pod); err != nil {
		t.Fatal(err)
	}

	// error should be waiting for instance to be running
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if qerr, ok := errors.IsRequeueError(err); !ok {
			t.Error("Expected requeue error, got:", err)
		} else if !strings.Contains(qerr.Error(), "not yet running") {
			t.Error("Expected waiting for desktop running, got:", qerr)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// set the container to running
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: desktop.GetName(), Namespace: desktop.GetNamespace()}, pod); err != nil {
		t.Fatal(err)
	}
	pod.Status = corev1.PodStatus{
		Phase: corev1.PodRunning,
		ContainerStatuses: []corev1.ContainerStatus{
			{
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
			},
		},
	}
	if err := r.client.Status().Update(context.TODO(), pod); err != nil {
		t.Fatal(err)
	}

	// reconciler should be waiting for a volume
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if qerr, ok := errors.IsRequeueError(err); !ok {
			t.Error("Expected requeue error, got:", err)
		} else if !strings.Contains(qerr.Error(), "volume provisioned yet") {
			t.Error("Expected waiting for volume, got:", qerr)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// get the pvc and update it with a volume name
	pvc, err := r.getPVCForInstance(cluster, desktop)
	if err != nil {
		t.Fatal(err)
	}
	pvc.Spec.VolumeName = "test-volume-name"
	if err := r.client.Update(context.TODO(), pvc); err != nil {
		t.Fatal(err)
	}
	// Create a volume backing the PVC
	pv := &corev1.PersistentVolume{}
	pv.Name = "test-volume-name"
	if err := r.client.Create(context.TODO(), pv); err != nil {
		t.Fatal(err)
	}

	// Reconcile should complete successfully
	// TODO: Check created resources, probably use envtest
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		t.Error("Expected reconcile to finish completely, got:", err)
	}

	// mock a deletion
	now := metav1.Now()
	desktop.SetDeletionTimestamp(&now)

	// should wait for pod to be gone
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if qerr, ok := errors.IsRequeueError(err); !ok {
			t.Error("Expected requeue error, got:", err)
		} else if !strings.Contains(qerr.Error(), "still terminating") {
			t.Error("Error should be pod still terminating, got:", err)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// delete the pod
	if err := r.client.Delete(context.TODO(), pod); err != nil {
		if client.IgnoreNotFound(err) != nil {
			t.Fatal(err)
		}
	}

	// should be waiting for the pvc to be terminated
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		if qerr, ok := errors.IsRequeueError(err); !ok {
			t.Error("Expected requeue error, got:", err)
		} else if !strings.Contains(qerr.Error(), "PVC is still being terminated") {
			t.Error("Error should be pvc still terminating, got:", err)
		}
	} else if err == nil {
		t.Error("Expected error got nil")
	}

	// delete the pvc
	if err := r.client.Delete(context.TODO(), pvc); err != nil {
		if client.IgnoreNotFound(err) != nil {
			t.Fatal(err)
		}
	}

	// Reconcile should complete
	// TODO: Again should check present resources
	if err := r.Reconcile(context.TODO(), testLogger, desktop); err != nil {
		t.Error("Expected reconcile to finish completely, got:", err)
	}
}
