package desktop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/pki"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler implements a reconciler for Desktop instance related resources.
type Reconciler struct {
	resources.DesktopReconciler

	client client.Client
	scheme *runtime.Scheme
}

var _ resources.DesktopReconciler = &Reconciler{}

var userdataReclaimFinalizer = "kvdi.io/userdata-reclaim"

// Global map of ticker routines. The UID of the desktop is placed as a key to
// avoid duplicate goroutines spawning.
var tickerRoutines = make(map[types.UID]struct{})

// New returns a new Desktop reconciler
func New(c client.Client, s *runtime.Scheme) *Reconciler {
	return &Reconciler{client: c, scheme: s}
}

// Reconcile ensures the required resources for a desktop session.
func (f *Reconciler) Reconcile(reqLogger logr.Logger, instance *v1alpha1.Desktop) error {
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
	if err := reconcile.Service(reqLogger, f.client, newServiceForCR(cluster, instance)); err != nil {
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

	// Set up a temporary connection to the secrets engine
	secretsEngine := secrets.GetSecretEngine(cluster)
	if err := secretsEngine.Setup(f.client, cluster); err != nil {
		return err
	}
	defer func() {
		if err := secretsEngine.Close(); err != nil {
			reqLogger.Error(err, "Error cleaning up secrets engine")
		}
	}()

	// ensure a certificate for novnc over mtls
	if err := pki.New(f.client, cluster, secretsEngine).ReconcileDesktop(reqLogger, instance, desktopSvc.Spec.ClusterIP); err != nil {
		return err
	}

	// If a secret was pre-created by the API for extra environment variables, fetch its name
	var secretName string
	if envTemplates := template.GetEnvTemplates(); len(envTemplates) > 0 {
		secretList := &corev1.SecretList{}
		if err := f.client.List(context.TODO(), secretList, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{v1.DesktopNameLabel: instance.GetName()}); err != nil {
			return err
		}
		if len(secretList.Items) == 0 {
			return errors.New("Could not find env secret for this desktop instance")
		}
		var secret *corev1.Secret
		for _, s := range secretList.Items {
			if strings.HasPrefix(s.GetName(), instance.GetUser()) {
				secret = &s
				break
			}
		}
		if secret == nil {
			return errors.New("Could not find env secret for this desktop instance")
		}
		if refs := secret.GetOwnerReferences(); len(refs) == 0 {
			secret.OwnerReferences = instance.OwnerReferences()
			if err := f.client.Update(context.TODO(), secret); err != nil {
				return err
			}
		}
		secretName = secret.GetName()
	}

	// ensure the pod
	if _, err := reconcile.Pod(reqLogger, f.client, newDesktopPodForCR(cluster, template, instance, secretName)); err != nil {
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

	// start a timer to kill the desktop if max session length is set
	if dur := cluster.GetMaxSessionLength(); dur != 0 {
		if _, ok := tickerRoutines[instance.GetUID()]; ok {
			// we already have a goroutine running, we are done here
			return nil
		}
		tickerRoutines[instance.GetUID()] = struct{}{}
		go func() {
			reqLogger.Info("Starting session timer for desktop instance.")

			// make sure to clean the global map on return
			defer func() { delete(tickerRoutines, instance.GetUID()) }()

			// define the namespaced name and setup tickers
			nn := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}
			sessTicker := time.NewTicker(dur)
			pollTicker := time.NewTicker(time.Duration(10) * time.Second)

			// listen on the ticker channels
			for {
				select {

				case <-sessTicker.C:
					// the desktop session has expired
					reqLogger.Info("Desktop session has expired, destroying instance")
					if err := f.client.Delete(context.TODO(), instance); err != nil {
						if client.IgnoreNotFound(err) != nil {
							reqLogger.Error(err, fmt.Sprintf("Error destroying desktop instance: %s", err.Error()))
						}
					}
					return

				case <-pollTicker.C:
					// return if desktop has been deleted
					if err := f.client.Get(context.TODO(), nn, &v1alpha1.Desktop{}); err != nil {
						if client.IgnoreNotFound(err) == nil {
							reqLogger.Info("Desktop instance has been deleted, stopping session poll")
							return
						}
						reqLogger.Error(err, fmt.Sprintf("Error polling desktop instance: %s", err.Error()))
						// retry on next loop
					}

				}
			}
		}()
	}

	return nil
}

func (f *Reconciler) updateNonRunningStatusAndRequeue(instance *v1alpha1.Desktop, pod *corev1.Pod, msg string) error {
	instance.Status.Running = false
	instance.Status.PodPhase = pod.Status.Phase
	if err := f.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return errors.NewRequeueError(msg, 3)
}

func (f *Reconciler) ensureFinalizers(reqLogger logr.Logger, instance *v1alpha1.Desktop) error {
	if !common.StringSliceContains(instance.GetFinalizers(), userdataReclaimFinalizer) {
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

func (f *Reconciler) runFinalizers(reqLogger logr.Logger, instance *v1alpha1.Desktop) error {
	var updated bool
	if common.StringSliceContains(instance.GetFinalizers(), userdataReclaimFinalizer) {
		if err := f.reclaimVolumes(reqLogger, instance); err != nil {
			return err
		}
		instance.SetFinalizers(common.StringSliceRemove(instance.GetFinalizers(), userdataReclaimFinalizer))
		updated = true
	}
	if updated {
		return f.client.Update(context.TODO(), instance)
	}
	return nil
}
