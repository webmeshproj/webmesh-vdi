// Package oidc contains an AuthProvider implementation backed by OpenID/Oauth.
package oidc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	gooidc "github.com/coreos/go-oidc"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthProvider implements an auth provider that uses an OIDC provider as the
// authentication backend. Access to groups provided in the claims is supplied
// through annotations on VDIRoles.
type AuthProvider struct {
	common.AuthProvider

	// k8s client
	client client.Client
	// our cluster instance
	cluster *v1alpha1.VDICluster
	// the secrets engine where we store our passwd
	secrets *secrets.SecretEngine
	// the oauth2 configuration
	oauthCfg oauth2.Config
	// verifier for verifying id tokens
	verifier *gooidc.IDTokenVerifier
	// the context containing our http client
	ctx context.Context
}

// Blank assignment to make sure AuthProvider satisfies the interface.
var _ common.AuthProvider = &AuthProvider{}

// New returns a new OIDC AuthProvider.
func New() common.AuthProvider {
	return &AuthProvider{}
}

// Setup implements the AuthProvider interface and sets a local reference to the
// k8s client and vdi cluster. It then configures oauth2/oidc for serving authentication
// requests.
func (a *AuthProvider) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	a.client = c
	a.cluster = cluster
	a.secrets = secrets.GetSecretEngine(cluster)

	if err := a.secrets.Setup(c, cluster); err != nil {
		return err
	}

	clientIDKey := a.cluster.GetOIDCClientIDKey()
	clientSecretKey := a.cluster.GetOIDCClientSecretKey()
	oidcSecrets, err := common.GetAuthSecrets(a.client, a.cluster, a.secrets, clientIDKey, clientSecretKey)
	if err != nil {
		return err
	}

	httpClient := &http.Client{}
	if strings.HasPrefix(a.cluster.GetOIDCIssuerURL(), "https") {
		caCert, err := a.cluster.GetOIDCCA()
		if err != nil {
			return err
		}
		var caCertPool *x509.CertPool
		if caCert != nil {
			caCertPool = x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
		}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: a.cluster.GetOIDCInsecureSkipVerify(),
				RootCAs:            caCertPool,
			},
		}
	}

	a.ctx = gooidc.ClientContext(context.Background(), httpClient)
	provider, err := gooidc.NewProvider(a.ctx, a.cluster.GetOIDCIssuerURL())
	if err != nil {
		return err
	}

	a.oauthCfg = oauth2.Config{
		ClientID:     oidcSecrets[clientIDKey],
		ClientSecret: oidcSecrets[clientSecretKey],
		RedirectURL:  a.cluster.GetOIDCRedirectURL(),
		Endpoint:     provider.Endpoint(),
		Scopes:       a.cluster.GetOIDCScopes(),
	}
	a.verifier = provider.Verifier(&gooidc.Config{ClientID: oidcSecrets[clientIDKey]})

	return nil
}

// Reconcile just makes sure that we have everything needed to perform an OIDC flow.
// The generated admin password is ignored for now in place of configuring admin groups.
func (a *AuthProvider) Reconcile(reqLogger logr.Logger, c client.Client, cluster *v1alpha1.VDICluster, adminPass string) error {
	return a.Setup(c, cluster)
}

// Close just returns nil as connections are not persistent
func (a *AuthProvider) Close() error {
	return nil
}
