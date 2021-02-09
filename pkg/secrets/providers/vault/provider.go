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
	"encoding/base64"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets/common"

	"github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var vaultLogger = logf.Log.WithName("vault_secrets")

// Provider implements a SecretsProvider that matches secret names to keys in
// vault.
type Provider struct {
	common.SecretsProvider

	crConfig    *v1alpha1.VaultConfig
	vaultConfig *api.Config
	client      *api.Client
	stopCh      chan struct{}
	getAuth     func(*v1alpha1.VaultConfig, *api.Config) (*api.Secret, error)
}

// Blank assignmnt to make sure Provider satisfies the SecretsProvider
// interface.
var _ common.SecretsProvider = &Provider{}

// New returns a new Provider.
func New() *Provider {
	return &Provider{
		getAuth: getK8sAuth,
	}
}

// Setup will set configurations then make sure we are able to read a k8s token
// and gain vault access with it. If authentication succeeds, a loop is spawned
// to keep the token fresh.
func (p *Provider) Setup(client client.Client, cluster *v1alpha1.VDICluster) error {
	var err error
	p.crConfig = cluster.Spec.Secrets.Vault
	p.vaultConfig, err = buildConfig(p.crConfig)
	if err != nil {
		return err
	}
	p.client, err = api.NewClient(p.vaultConfig)
	if err != nil {
		return err
	}
	auth, err := p.getAuth(p.crConfig, p.vaultConfig)
	if err != nil {
		return err
	}
	p.client.SetToken(auth.Auth.ClientToken)
	if p.stopCh == nil {
		p.stopCh = make(chan struct{})
		go p.runTokenRefreshLoop(auth)
	}
	return nil
}

// Close signals the stop channel if it's been created, and revokes the token
// if there is a client configured.
func (p *Provider) Close() error {
	if p.stopCh != nil {
		p.stopCh <- struct{}{}
	}
	if p.client != nil {
		// RevokeSelf ignores its parameters and uses the client's set token.
		return p.client.Auth().Token().RevokeSelf("")
	}
	return nil
}

// buildConfig builds a vault API configuration.
func buildConfig(conf *v1alpha1.VaultConfig) (*api.Config, error) {
	var caCert string
	if conf.CACertBase64 != "" {
		caCertBytes, err := base64.StdEncoding.DecodeString(conf.CACertBase64)
		if err != nil {
			return nil, err
		}
		caCert = string(caCertBytes)
	}
	config := api.DefaultConfig()
	if err := config.ConfigureTLS(&api.TLSConfig{
		CACert:        caCert,
		TLSServerName: conf.TLSServerName,
		Insecure:      conf.Insecure,
	}); err != nil {
		return nil, err
	}
	config.Address = conf.Address
	return config, nil
}
