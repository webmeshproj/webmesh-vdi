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

func ReconcileCertificate(reqLogger logr.Logger, c client.Client, cert *cm.Certificate, wait bool) error {
	if err := util.SetCreationSpecAnnotation(&cert.ObjectMeta, cert); err != nil {
		return err
	}

	found := &cm.Certificate{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: cert.Name, Namespace: cert.Namespace}, found); err != nil {
		// Return API error
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		reqLogger.Info("Creating new certificate", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)
		// Create the certificate
		if err := c.Create(context.TODO(), cert); err != nil {
			return err
		}
		if wait {
			return errors.NewRequeueError("Requeueing status check for new certificate", 3)
		}
		return nil
	}

	// Check the found certificate spec
	if !util.CreationSpecsEqual(cert.ObjectMeta, found.ObjectMeta) {
		// We need to update the certificate
		found.Spec = cert.Spec
		if err := c.Update(context.TODO(), found); err != nil {
			return err
		}
	}

	if wait {
		for _, condition := range found.Status.Conditions {
			if condition.Type == cm.CertificateConditionReady {
				if condition.Status == cmmeta.ConditionTrue {
					return nil
				}
			}
		}
		return errors.NewRequeueError("Certificate is not ready yet", 3)
	}

	return nil
}
