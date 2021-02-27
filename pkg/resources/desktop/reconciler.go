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
	"fmt"
	"strings"
	"time"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"

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
func (f *Reconciler) Reconcile(ctx context.Context, reqLogger logr.Logger, instance *desktopsv1.Session) error {
	if instance.GetDeletionTimestamp() != nil {
		return f.runFinalizers(ctx, reqLogger, instance)
	}

	reqLogger.Info("Retrieving template and cluster for session")

	template, err := instance.GetTemplate(f.client)
	if err != nil {
		return err
	}
	cluster, err := instance.GetVDICluster(f.client)
	if err != nil {
		return err
	}

	resourceNamespacedName := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}

	var userdataVol string
	// create a PV for the user if we need to
	if selector := cluster.GetUserdataSelector(); selector != nil && selector.IsValid() {
		reqLogger.Info("Cluster has userdataSelector, searching for user PVC")
		userdataVol, err = f.locateUserdataPVC(ctx, reqLogger, instance, selector)
		if err != nil {
			return err
		}
	} else if cluster.GetUserdataVolumeSpec() != nil {
		reqLogger.Info("Cluster has userdataSpec, reconciling volumes")
		if err := f.reconcileVolumes(ctx, reqLogger, cluster, instance); err != nil {
			return err
		}
		userdataVol = cluster.GetUserdataVolumeName(instance.GetUser())
	}

	// create a service in front of the desktop (so we can pre-allocate an IP that resolves to the pod)
	reqLogger.Info("Reconciling service for the desktop session")
	if err := reconcile.Service(ctx, reqLogger, f.client, newServiceForCR(cluster, instance)); err != nil {
		return err
	}

	// get the service IP
	desktopSvc := &corev1.Service{}
	if err := f.client.Get(ctx, resourceNamespacedName, desktopSvc); err != nil {
		return err
	}

	if desktopSvc.Spec.ClusterIP == "" || desktopSvc.Spec.ClusterIP == "None" {
		return errors.NewRequeueError("Desktop service has not yet been assigned an IP", 2)
	}

	// Set up a temporary connection to the secrets engine
	reqLogger.Info("Generating mTLS certificate for the session proxy")
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
	if template.HasManagedEnvSecret() {
		reqLogger.Info("Session has secret environment variables from user, retrieving")
		secretList := &corev1.SecretList{}
		if err := f.client.List(ctx, secretList, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{v1.DesktopNameLabel: instance.GetName()}); err != nil {
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
			if err := f.client.Update(ctx, secret); err != nil {
				return err
			}
		}
		secretName = secret.GetName()
	}

	// ensure the pod
	reqLogger.Info("Reconciling pod for session")
	if _, err := reconcile.Pod(ctx, reqLogger, f.client, newDesktopPodForCR(cluster, template, instance, secretName, userdataVol)); err != nil {
		return err
	}

	// Wait for the desktop to be ready
	desktopPod := &corev1.Pod{}
	nn := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}
	if err := f.client.Get(ctx, nn, desktopPod); err != nil {
		return err
	}

	if desktopPod.Status.Phase != corev1.PodRunning {
		return f.updateNonRunningStatusAndRequeue(ctx, instance, desktopPod, "Desktop pod is not in running phase")
	}
	for _, status := range desktopPod.Status.ContainerStatuses {
		if status.State.Running == nil {
			return f.updateNonRunningStatusAndRequeue(ctx, instance, desktopPod, "Desktop instance is not yet running")
		}
	}

	if (cluster.GetUserdataSelector() == nil || !cluster.GetUserdataSelector().IsValid()) && cluster.GetUserdataVolumeSpec() != nil {
		if err := f.reconcileUserdataMapping(ctx, reqLogger, cluster, instance); err != nil {
			return err
		}
		if err := f.ensureFinalizers(ctx, reqLogger, instance); err != nil {
			return err
		}
	}

	if !instance.Status.Running {
		instance.Status.PodPhase = desktopPod.Status.Phase
		instance.Status.Running = true
		if err := f.client.Status().Update(ctx, instance); err != nil {
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
		go f.killOnSessionTimeout(reqLogger, instance, dur)
	}

	return nil
}

func (f *Reconciler) locateUserdataPVC(ctx context.Context, reqLogger logr.Logger, instance *desktopsv1.Session, selector *appv1.UserdataSelector) (string, error) {
	if selector.MatchName != "" {
		var pvc corev1.PersistentVolumeClaim
		nn := types.NamespacedName{
			Name:      strings.Replace(selector.MatchName, "${USERNAME}", instance.GetUser(), -1),
			Namespace: instance.GetNamespace(),
		}
		err := f.client.Get(ctx, nn, &pvc)
		if err != nil {
			return "", err
		}
		return pvc.GetName(), nil
	}
	if selector.MatchLabel != "" {
		var pvcList corev1.PersistentVolumeClaimList
		err := f.client.List(
			ctx, &pvcList,
			client.InNamespace(instance.GetNamespace()),
			client.MatchingLabels{selector.MatchLabel: instance.GetUser()},
		)
		if err != nil {
			return "", err
		}
		if len(pvcList.Items) == 0 {
			return "", fmt.Errorf("%s=%s did not return any PVCs in namespace %s",
				selector.MatchLabel, instance.GetUser(), instance.GetNamespace())
		}
		if len(pvcList.Items) > 1 {
			return "", fmt.Errorf("%s=%s returned multiple PVCs in namespace %s",
				selector.MatchLabel, instance.GetUser(), instance.GetNamespace())
		}
		return pvcList.Items[0].GetName(), nil
	}
	// Safeguard but would never fire if selector is pre-checked for validity
	return "", errors.New("Cannot use empty userdata selector")
}

func (f *Reconciler) killOnSessionTimeout(reqLogger logr.Logger, instance *desktopsv1.Session, dur time.Duration) {
	ctx := context.Background()

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
			if err := f.client.Delete(ctx, instance); err != nil {
				if client.IgnoreNotFound(err) != nil {
					reqLogger.Error(err, fmt.Sprintf("Error destroying desktop instance: %s", err.Error()))
				}
			}
			return

		case <-pollTicker.C:
			// return if desktop has been deleted
			if err := f.client.Get(ctx, nn, &desktopsv1.Session{}); err != nil {
				if client.IgnoreNotFound(err) == nil {
					reqLogger.Info("Desktop instance has been deleted, stopping session poll")
					return
				}
				reqLogger.Error(err, fmt.Sprintf("Error polling desktop instance: %s", err.Error()))
				// retry on next loop
			}

		}
	}
}

func (f *Reconciler) updateNonRunningStatusAndRequeue(ctx context.Context, instance *desktopsv1.Session, pod *corev1.Pod, msg string) error {
	instance.Status.Running = false
	instance.Status.PodPhase = pod.Status.Phase
	if err := f.client.Status().Update(ctx, instance); err != nil {
		return err
	}
	return errors.NewRequeueError(msg, 3)
}

func (f *Reconciler) ensureFinalizers(ctx context.Context, reqLogger logr.Logger, instance *desktopsv1.Session) error {
	if !common.StringSliceContains(instance.GetFinalizers(), userdataReclaimFinalizer) {
		instance.SetFinalizers(append(instance.GetFinalizers(), userdataReclaimFinalizer))
		if err := f.client.Update(ctx, instance); err != nil {
			return err
		}
		if err := f.client.Get(ctx, types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, instance); err != nil {
			return err
		}
	}
	return nil
}

func (f *Reconciler) runFinalizers(ctx context.Context, reqLogger logr.Logger, instance *desktopsv1.Session) error {
	var updated bool
	if common.StringSliceContains(instance.GetFinalizers(), userdataReclaimFinalizer) {
		if err := f.reclaimVolumes(reqLogger, instance); err != nil {
			return err
		}
		instance.SetFinalizers(common.StringSliceRemove(instance.GetFinalizers(), userdataReclaimFinalizer))
		updated = true
	}
	if updated {
		return f.client.Update(ctx, instance)
	}
	return nil
}
