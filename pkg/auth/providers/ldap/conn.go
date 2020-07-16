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

func (a *AuthProvider) bind() error {
	return a.conn.Bind(a.bindDN, a.bindPassw)
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
