package local

import (
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const passwdKey = "passwd"

func (l *LocalAuthProvider) Reconcile(reqLogger logr.Logger, c client.Client, cluster *v1alpha1.VDICluster, adminPass string) error {
	secretsEngine := secrets.GetSecretEngine(cluster)
	if err := secretsEngine.Setup(c, cluster); err != nil {
		return err
	}
	defer func() {
		if err := secretsEngine.Close(); err != nil {
			reqLogger.Error(err, "Error cleaning up secrets engine")
		}
	}()

	if _, err := secretsEngine.ReadSecret(passwdKey, false); err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return err
		}
		adminRole := cluster.GetAdminRole()
		hash, err := common.HashPassword(adminPass)
		if err != nil {
			return err
		}
		if err := secretsEngine.WriteSecret(passwdKey, []byte(fmt.Sprintf("admin:%s:%s\n", adminRole.GetName(), hash))); err != nil {
			return err
		}
	}
	return nil
}
