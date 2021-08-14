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

package vault

import (
	"net"
	"testing"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestVaultCore(t *testing.T) (*vault.Core, net.Listener, *appv1.VDICluster, string) {
	t.Helper()
	core, _, token := vault.TestCoreUnsealed(t)
	srvr, addr := http.TestServer(t, core)
	return core, srvr, &appv1.VDICluster{
		Spec: appv1.VDIClusterSpec{
			Secrets: &appv1.SecretsConfig{
				Vault: &appv1.VaultConfig{
					Address:     addr,
					Insecure:    true,
					SecretsPath: "secret/",
				},
			},
		},
	}, token
}

func mustSetupProvider(t *testing.T) (*Provider, net.Listener, *vault.Core) {
	t.Helper()

	core, srvr, cr, rootToken := newTestVaultCore(t)

	// create a provider and override get auth to return the token to the test core
	provider := New()
	provider.getAuth = func(*appv1.VaultConfig, *api.Config) (*api.Secret, error) {
		return &api.Secret{
			Auth: &api.SecretAuth{
				ClientToken:   rootToken,
				Renewable:     true,
				LeaseDuration: 120,
			},
		}, nil
	}
	if err := provider.Setup(fake.NewFakeClientWithScheme(runtime.NewScheme()), cr); err != nil {
		t.Fatal(err)
	}

	return provider, srvr, core
}
