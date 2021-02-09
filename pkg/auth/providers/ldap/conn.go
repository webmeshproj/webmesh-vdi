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

package ldap

import (
	"crypto/tls"
	"crypto/x509"

	ldapv3 "github.com/go-ldap/ldap/v3"
)

// connect creates a connection with the ldap server. It assumes the credentials
// are already present in the current interface.
func (a *AuthProvider) connect() (*ldapv3.Conn, error) {
	if a.cluster.IsUsingLDAPOverTLS() {
		return ldapv3.DialURL(a.cluster.GetLDAPURL(), ldapv3.DialWithTLSConfig(a.tlsConfig))
	}
	return ldapv3.DialURL(a.cluster.GetLDAPURL())
}

func (a *AuthProvider) bind(conn *ldapv3.Conn) error {
	return conn.Bind(a.bindDN, a.bindPassw)
}

func (a *AuthProvider) fetchAndSetBindCredentials() error {
	var err error
	a.bindDN, a.bindPassw, err = a.getCredentials()
	return err
}

func (a *AuthProvider) setTLSConfig() error {
	caCert, err := a.cluster.GetLDAPCA()
	if err != nil {
		return err
	}
	var caCertPool *x509.CertPool
	if caCert != nil {
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}
	a.tlsConfig = &tls.Config{
		InsecureSkipVerify: a.cluster.GetLDAPInsecureSkipVerify(),
		RootCAs:            caCertPool,
	}
	return nil
}
