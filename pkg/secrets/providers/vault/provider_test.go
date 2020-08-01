package vault

import (
	"net"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestVaultCore(t *testing.T) (*vault.Core, net.Listener, *v1alpha1.VDICluster, string) {
	t.Helper()
	core, _, token := vault.TestCoreUnsealed(t)
	srvr, addr := http.TestServer(t, core)
	return core, srvr, &v1alpha1.VDICluster{
		Spec: v1alpha1.VDIClusterSpec{
			Secrets: &v1alpha1.SecretsConfig{
				Vault: &v1alpha1.VaultConfig{
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
	provider.getAuth = func(*v1alpha1.VaultConfig, *api.Config) (*api.Secret, error) {
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
