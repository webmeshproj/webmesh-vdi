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

func ReconcileDeployment(reqLogger logr.Logger, c client.Client, deployment *appsv1.Deployment, wait bool) error {
	if err := util.SetCreationSpecAnnotation(&deployment.ObjectMeta, deployment); err != nil {
		return err
	}

	foundDeployment := &appsv1.Deployment{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the deployment
		reqLogger.Info("Creating new deployment", "Deployment.Name", deployment.Name, "Deployment.Namespace", deployment.Namespace)
		if err := c.Create(context.TODO(), deployment); err != nil {
			return err
		}
		if wait {
			return errors.NewRequeueError("Created new deployment with wait, requeing for status check", 3)
		}
		return nil
	}

	// Check the found deployment spec
	if !util.CreationSpecsEqual(deployment.ObjectMeta, foundDeployment.ObjectMeta) {
		// We need to update the deployment
		reqLogger.Info("Deployment annotation spec has changed, updating", "Deployment.Name", deployment.Name, "Deployment.Namespace", deployment.Namespace)
		foundDeployment.Spec = deployment.Spec
		if err := c.Update(context.TODO(), foundDeployment); err != nil {
			return err
		}
	}

	if wait {
		runningDeploy := &appsv1.Deployment{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, runningDeploy); err != nil {
			return err
		}
		if runningDeploy.Status.ReadyReplicas != *deployment.Spec.Replicas {
			return errors.NewRequeueError(fmt.Sprintf("Waiting for %s to be ready", deployment.Name), 3)
		}
	}

	return nil
}
