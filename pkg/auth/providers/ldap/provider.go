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

// Package ldap contains an AuthProvider implementation that uses a remote
// LDAP server for authentication.
package ldap

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"github.com/kvdi/kvdi/pkg/auth/common"
	"github.com/kvdi/kvdi/pkg/secrets"
)

// AuthProvider implements an auth provider that uses an LDAP server as the
// authentication backend. Access to groups in LDAP is supplied through annotations
// on VDIRoles.
type AuthProvider struct {
	// k8s client
	client client.Client
	// our cluster instance
	cluster *appv1.VDICluster
	// the secrets engine where we store our passwd
	secrets *secrets.SecretEngine
	// the user dn for binding to ldap
	bindDN string
	// the password for binding to ldap
	bindPassw string
	// a tls configuration if using TLS
	tlsConfig *tls.Config
	// the base DN for the connected LDAP server
	baseDN string
}

// Blank assignment to make sure AuthProvider satisfies the interface.
var _ common.AuthProvider = &AuthProvider{}

// New returns a new LDAPAuthProvider.
func New(s *secrets.SecretEngine) common.AuthProvider {
	return &AuthProvider{secrets: s}
}

// Setup implements the AuthProvider interface and sets a local reference to the
// k8s client and vdi cluster.
func (a *AuthProvider) Setup(c client.Client, cluster *appv1.VDICluster) error {
	a.client = c
	a.cluster = cluster

	var err error

	if err = a.fetchAndSetBindCredentials(); err != nil {
		return err
	}

	if a.cluster.IsUsingLDAPOverTLS() {
		if err = a.setTLSConfig(); err != nil {
			return err
		}
	}

	baseDnFields := make([]string, 0)
	for _, field := range strings.Split(a.bindDN, ",") {
		if strings.HasPrefix(strings.ToLower(field), "dc") {
			baseDnFields = append(baseDnFields, field)
		}
	}
	a.baseDN = strings.Join(baseDnFields, ",")

	// verify we can connect to the ldap server
	conn, err := a.connect()
	if err != nil {
		return err
	}

	// verify credentials work
	defer conn.Close()
	return a.bind(conn)
}

// Reconcile just makes sure that we are able to succesfully set up a connection.
// The generated admin password is ignored for now in place of configuring admin groups.
func (a *AuthProvider) Reconcile(ctx context.Context, reqLogger logr.Logger, c client.Client, cluster *appv1.VDICluster, adminPass string) error {
	return a.Setup(c, cluster)
}

// Close just returns nil as connections are not persistent
func (a *AuthProvider) Close() error {
	return nil
}
