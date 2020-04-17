package reconcile

import (
	"context"

	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileClusterIssuer(reqLogger logr.Logger, c client.Client, issuer *cm.ClusterIssuer, wait bool) error {
	if err := util.SetCreationSpecAnnotation(&issuer.ObjectMeta, issuer); err != nil {
		return err
	}

	found := &cm.ClusterIssuer{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: issuer.Name, Namespace: issuer.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Create the issuer
		reqLogger.Info("Creating new cluster issuer", "ClusterIssuer.Name", issuer.Name, "ClusterIssuer.Namespace", issuer.Namespace)
		if err := c.Create(context.TODO(), issuer); err != nil {
			return err
		}
		if wait {
			return errors.NewRequeueError("Requeueing status check for new issuer", 3)
		}
		return nil
	}

	// Check the found certificate spec
	if !util.CreationSpecsEqual(issuer.ObjectMeta, found.ObjectMeta) {
		// We need to update the certificate
		found.Spec = issuer.Spec
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
	}

	if wait {
		for _, condition := range found.Status.Conditions {
			if condition.Type == cm.IssuerConditionReady {
				if condition.Status == cmmeta.ConditionTrue {
					return nil
				}
			}
		}
		return errors.NewRequeueError("Issuer is not ready yet", 3)
	}

	return nil
}
