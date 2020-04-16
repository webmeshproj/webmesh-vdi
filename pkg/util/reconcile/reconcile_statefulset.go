package reconcile

import (
	"context"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileStatefulSet(reqLogger logr.Logger, c client.Client, ss *appsv1.StatefulSet, wait bool) error {
	if err := util.SetCreationSpecAnnotation(&ss.ObjectMeta, ss); err != nil {
		return err
	}

	foundStatefulSet := &appsv1.StatefulSet{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, foundStatefulSet); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the ss
		reqLogger.Info("Creating new statefulset", "StatefulSet.Name", ss.Name, "StatefulSet.Namespace", ss.Namespace)
		if err := c.Create(context.TODO(), ss); err != nil {
			return err
		}
		if wait {
			return errors.NewRequeueError("Created new statefulset with wait, requeing for status check", 3)
		}
		return nil
	}

	// Check the found ss spec
	if !util.CreationSpecsEqual(ss.ObjectMeta, foundStatefulSet.ObjectMeta) {
		// We need to update the ss
		reqLogger.Info("StatefulSet annotation spec has changed, updating", "StatefulSet.Name", ss.Name, "StatefulSet.Namespace", ss.Namespace)
		foundStatefulSet.Spec = ss.Spec
		if err := c.Update(context.TODO(), foundStatefulSet); err != nil {
			return err
		}
	}

	if wait {
		runningDeploy := &appsv1.StatefulSet{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, runningDeploy); err != nil {
			return err
		}
		if runningDeploy.Status.ReadyReplicas != *ss.Spec.Replicas {
			return errors.NewRequeueError(fmt.Sprintf("Waiting for %s to be ready", ss.Name), 3)
		}
	}

	return nil
}
