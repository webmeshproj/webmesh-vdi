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
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var replicas int32 = 1

func newFakeDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-deployment",
			Namespace: "fake-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
	}
}

func TestReconcileDeployment(t *testing.T) {
	c := getFakeClient(t)
	deployment := newFakeDeployment()
	nn := types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}

	if err := Deployment(testLogger, c, deployment, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	if err := c.Get(context.TODO(), nn, deployment); err != nil {
		t.Error("Expected deployment to exist, got:", err)
	}

	if err := Deployment(testLogger, c, deployment, false); err != nil {
		t.Error("Expected no error, got", err)
	}

	deployment.Status = appsv1.DeploymentStatus{
		ReadyReplicas: 0,
	}
	c.Status().Update(context.TODO(), deployment)

	if err := Deployment(testLogger, c, deployment, true); err == nil {
		t.Error("Expected requeue error, got nil")
	} else if _, ok := errors.IsRequeueError(err); !ok {
		t.Error("Expected requeue error, got:", err)
	}

	c.Get(context.TODO(), nn, deployment)
	deployment.Status = appsv1.DeploymentStatus{
		ReadyReplicas: 1,
	}
	if err := c.Status().Update(context.TODO(), deployment); err != nil {
		t.Fatal(err)
	}
	if err := Deployment(testLogger, c, deployment, true); err != nil {
		t.Error("Expected no error, got", err)
	}
}
