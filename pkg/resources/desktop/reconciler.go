package desktop

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesktopReconciler struct {
	resources.DesktopReconciler

	client client.Client
	scheme *runtime.Scheme
}

var _ resources.DesktopReconciler = &DesktopReconciler{}

var userdataReclaimFinalizer = "kvdi.io/userdata-reclaim"

// New returns a new Desktop reconciler
func New(c client.Client, s *runtime.Scheme) resources.DesktopReconciler {
	return &DesktopReconciler{client: c, scheme: s}
}

func (f *DesktopReconciler) Reconcile(reqLogger logr.Logger, instance *v1alpha1.Desktop) error {
	if instance.GetDeletionTimestamp() != nil {
		return f.runFinalizers(reqLogger, instance)
	}

	template, err := instance.GetTemplate(f.client)
	if err != nil {
		return err
	}
	cluster, err := instance.GetVDICluster(f.client)
	if err != nil {
		return err
	}

	resourceNamespacedName := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}

	// create a PV for the user if we need to
	if cluster.GetUserdataVolumeSpec() != nil {
		if err := f.reconcileVolumes(reqLogger, cluster, instance); err != nil {
			return err
		}
	}

	// create a service in front of the desktop (so we can pre-allocate an IP that resolves to the pod)
	if err := reconcile.ReconcileService(reqLogger, f.client, newServiceForCR(cluster, instance)); err != nil {
		return err
	}

	// get the service IP
	desktopSvc := &corev1.Service{}
	if err := f.client.Get(context.TODO(), resourceNamespacedName, desktopSvc); err != nil {
		return err
	}

	if desktopSvc.Spec.ClusterIP == "" || desktopSvc.Spec.ClusterIP == "None" {
		return errors.NewRequeueError("Desktop service has not yet been assigned an IP", 2)
	}

	// ensure a certificate for novnc over mtls
	if err := reconcile.ReconcileCertificate(reqLogger, f.client, newDesktopProxyCert(cluster, instance, desktopSvc.Spec.ClusterIP), true); err != nil {
		return err
	}

	// ensure the pod
	if _, err := reconcile.ReconcilePod(reqLogger, f.client, newDesktopPodForCR(cluster, template, instance)); err != nil {
		return err
	}

	// Wait for the desktop to be ready
	desktopPod := &corev1.Pod{}
	nn := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}
	if err := f.client.Get(context.TODO(), nn, desktopPod); err != nil {
		return err
	}

	if desktopPod.Status.Phase != corev1.PodRunning {
		return f.updateNonRunningStatusAndRequeue(instance, desktopPod, "Desktop pod is not in running phase")
	}
	for _, status := range desktopPod.Status.ContainerStatuses {
		if status.State.Running == nil {
			return f.updateNonRunningStatusAndRequeue(instance, desktopPod, "Desktop instance is not yet running")
		}
	}

	if cluster.GetUserdataVolumeSpec() != nil {
		if err := f.reconcileUserdataMapping(reqLogger, cluster, instance); err != nil {
			return err
		}
		if err := f.ensureFinalizers(reqLogger, instance); err != nil {
			return err
		}
	}

	if !instance.Status.Running {
		instance.Status.PodPhase = desktopPod.Status.Phase
		instance.Status.Running = true
		if err := f.client.Status().Update(context.TODO(), instance); err != nil {
			return err
		}
	}

	return nil
}

func (f *DesktopReconciler) updateNonRunningStatusAndRequeue(instance *v1alpha1.Desktop, pod *corev1.Pod, msg string) error {
	instance.Status.Running = false
	instance.Status.PodPhase = pod.Status.Phase
	if err := f.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return errors.NewRequeueError(msg, 3)
}

func (f *DesktopReconciler) ensureFinalizers(reqLogger logr.Logger, instance *v1alpha1.Desktop) error {
	if !util.StringSliceContains(instance.GetFinalizers(), userdataReclaimFinalizer) {
		instance.SetFinalizers(append(instance.GetFinalizers(), userdataReclaimFinalizer))
		if err := f.client.Update(context.TODO(), instance); err != nil {
			return err
		}
		if err := f.client.Get(context.TODO(), types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, instance); err != nil {
			return err
		}
	}
	return nil
}

func (f *DesktopReconciler) runFinalizers(reqLogger logr.Logger, instance *v1alpha1.Desktop) error {
	var updated bool
	if util.StringSliceContains(instance.GetFinalizers(), userdataReclaimFinalizer) {
		if err := f.reclaimVolumes(reqLogger, instance); err != nil {
			return err
		}
		instance.SetFinalizers(util.StringSliceRemove(instance.GetFinalizers(), userdataReclaimFinalizer))
		updated = true
	}
	if updated {
		return f.client.Update(context.TODO(), instance)
	}
	return nil
}
