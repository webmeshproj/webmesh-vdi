package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/resources"
	"github.com/tinyzimmer/kvdi/pkg/util/reconcile"
	rdbutil "github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RethinkDBReconciler struct {
	resources.VDIReconciler

	client client.Client
	scheme *runtime.Scheme
}

var _ resources.VDIReconciler = &RethinkDBReconciler{}

// New returns a new rethinkdb reconciler
func New(c client.Client, s *runtime.Scheme) resources.VDIReconciler {
	return &RethinkDBReconciler{client: c, scheme: s}
}

func (r *RethinkDBReconciler) Reconcile(reqLogger logr.Logger, instance *v1alpha1.VDICluster) error {
	if err := reconcile.ReconcileCertificate(reqLogger, r.client, newDBCertForCR(instance), true); err != nil {
		return err
	}
	if err := reconcile.ReconcileCertificate(reqLogger, r.client, newMgrClientCertForCR(instance), true); err != nil {
		return err
	}
	if err := reconcile.ReconcileService(reqLogger, r.client, newServiceForCR(instance)); err != nil {
		return err
	}
	if err := reconcile.ReconcileService(reqLogger, r.client, newProxyServiceForCR(instance)); err != nil {
		return err
	}
	if err := reconcile.ReconcileStatefulSet(reqLogger, r.client, newStatefulSetForCR(instance), true); err != nil {
		return err
	}
	if err := r.reconcileProxy(reqLogger, instance); err != nil {
		return err
	}

	adminPass, err := r.reconcileAdminSecret(reqLogger, instance)
	if err != nil {
		return err
	}

	sess, err := rdbutil.NewFromSecret(r.client, rdbutil.RDBAddrForCR(instance), instance.GetRethinkDBClientName(), instance.GetCoreNamespace())
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.Migrate(adminPass, *instance.GetRethinkDBReplicas(), *instance.GetRethinkDBShards(), instance.AnonymousAllowed())
}
