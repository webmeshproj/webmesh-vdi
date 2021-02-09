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

package reconcile

import (
	"context"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment reconciles a deployment with the cluster and opionally
// returns a requeue error if it isn't fully running yet.
func Deployment(reqLogger logr.Logger, c client.Client, deployment *appsv1.Deployment, wait bool) error {
	if err := k8sutil.SetCreationSpecAnnotation(&deployment.ObjectMeta, deployment); err != nil {
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
	if !k8sutil.CreationSpecsEqual(deployment.ObjectMeta, foundDeployment.ObjectMeta) {
		// We need to update the deployment
		reqLogger.Info("Deployment annotation spec has changed, updating", "Deployment.Name", deployment.Name, "Deployment.Namespace", deployment.Namespace)
		foundDeployment.Spec = deployment.Spec
		foundDeployment.SetAnnotations(deployment.GetAnnotations())
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
