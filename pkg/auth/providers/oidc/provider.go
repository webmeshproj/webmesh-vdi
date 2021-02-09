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

// Package oidc contains an AuthProvider implementation backed by OpenID/Oauth.
package oidc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
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
	cluster *appv1.VDICluster
	// the secrets engine where we store our passwd
	secrets *secrets.SecretEngine
	// the oauth2 configuration
	oauthCfg oauth2.Config
	// verifier for verifying id tokens
	verifier *gooidc.IDTokenVerifier
	// the url that can be used for exchanging refresh tokens
	tokenURL string
	// the context containing our http client
	ctx context.Context
	// the client id
	clientID string
	// the client secret
	clientSecret string
}

// Blank assignment to make sure AuthProvider satisfies the interface.
var _ common.AuthProvider = &AuthProvider{}

// New returns a new OIDC AuthProvider.
func New(s *secrets.SecretEngine) common.AuthProvider {
	return &AuthProvider{secrets: s}
}

// Setup implements the AuthProvider interface and sets a local reference to the
// k8s client and vdi cluster. It then configures oauth2/oidc for serving authentication
// requests.
func (a *AuthProvider) Setup(c client.Client, cluster *appv1.VDICluster) error {
	a.client = c
	a.cluster = cluster

	clientIDKey := a.cluster.GetOIDCClientIDKey()
	clientSecretKey := a.cluster.GetOIDCClientSecretKey()
	oidcSecrets, err := common.GetAuthSecrets(a.client, a.cluster, a.secrets, clientIDKey, clientSecretKey)
	if err != nil {
		return err
	}

	a.clientID = oidcSecrets[clientIDKey]
	a.clientSecret = oidcSecrets[clientSecretKey]

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

	a.tokenURL = provider.Endpoint().TokenURL

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
func (a *AuthProvider) Reconcile(ctx context.Context, reqLogger logr.Logger, c client.Client, cluster *appv1.VDICluster, adminPass string) error {
	return a.Setup(c, cluster)
}

// Close just returns nil as connections are not persistent
func (a *AuthProvider) Close() error {
	return nil
}
