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

package local

import (
	"context"
	"fmt"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"github.com/kvdi/kvdi/pkg/util/common"
	"github.com/kvdi/kvdi/pkg/util/errors"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const passwdKey = "passwd"

// Reconcile prepares the resources required to use the local authentication driver.
func (l *AuthProvider) Reconcile(ctx context.Context, reqLogger logr.Logger, c client.Client, cluster *appv1.VDICluster, adminPass string) error {
	if _, err := l.secrets.ReadSecret(passwdKey, false); err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return err
		}
		adminRole := cluster.GetAdminRole()
		hash, err := common.HashPassword(adminPass)
		if err != nil {
			return err
		}
		if err := l.secrets.WriteSecret(passwdKey, []byte(fmt.Sprintf("admin:%s:%s\n", adminRole.GetName(), hash))); err != nil {
			return err
		}
	}
	return nil
}
