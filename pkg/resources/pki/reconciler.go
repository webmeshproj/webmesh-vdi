package pki

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PKIReconciler struct {
	resources.VDIReconciler

	client client.Client
	scheme *runtime.Scheme
}

var _ resources.VDIReconciler = &PKIReconciler{}

// New returns a new pki reconciler
func New(c client.Client, s *runtime.Scheme) resources.VDIReconciler {
	return &PKIReconciler{client: c, scheme: s}
}

func (p *PKIReconciler) Reconcile(reqLogger logr.Logger, instance *v1alpha1.VDICluster) error {
	if err := reconcile.ReconcileClusterIssuer(reqLogger, p.client, newSignerForCR(instance), true); err != nil {
		return err
	}
	if err := reconcile.ReconcileCertificate(reqLogger, p.client, newCAForCR(instance), true); err != nil {
		return err
	}
	if err := reconcile.ReconcileClusterIssuer(reqLogger, p.client, newIssuerForCR(instance), true); err != nil {
		return err
	}
	return nil
}
