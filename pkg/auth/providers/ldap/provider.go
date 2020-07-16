package ldap

import (
	"crypto/tls"
	"strings"
	"sync"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	ldapv3 "github.com/go-ldap/ldap/v3"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const userFilter = "(uid=%s)"
const groupUsersFilter = "(memberOf=%s)"

// secretsLog is the logr interface for the secrets engine
var ldapLog = logf.Log.WithName("ldap")

var userAttrs = []string{"cn", "dn", "memberOf", "accountStatus"}

// AuthProvider implements an auth provider that uses an LDAP server as the
// authentication backend. Access to groups in LDAP is supplied through annotations
// on VDIRoles.
type AuthProvider struct {
	v1alpha1.AuthProvider

	// k8s client
	client client.Client
	// our cluster instance
	cluster *v1alpha1.VDICluster
	// the secrets engine where we store our passwd
	secrets *secrets.SecretEngine
	// the user dn for binding to ldap
	bindDN string
	// the password for binding to ldap
	bindPassw string
	// a tls configuration if using TLS
	tlsConfig *tls.Config
	// the underlying ldap connection
	conn *ldapv3.Conn
	// the base DN for the connected LDAP server
	baseDN string
	// mutex for local locking
	mux sync.Mutex
}

// Blank assignment to make sure AuthProvider satisfies the interface.
var _ v1alpha1.AuthProvider = &AuthProvider{}

// New returns a new LDAPAuthProvider.
func New() v1alpha1.AuthProvider {
	return &AuthProvider{}
}

// Setup implements the AuthProvider interface and sets a local reference to the
// k8s client and vdi cluster.
func (a *AuthProvider) Setup(c client.Client, cluster *v1alpha1.VDICluster) error {
	a.client = c
	a.cluster = cluster
	a.secrets = secrets.GetSecretEngine(cluster)

	var err error

	if err = a.secrets.Setup(c, cluster); err != nil {
		return err
	}

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

	if a.conn == nil {
		a.conn, err = a.connect()
		if err != nil {
			return err
		}
		err = a.bind()
	}

	return err
}

// Reconcile just makes sure that we are able to succesfully set up a connection.
// The generated admin password is ignored for now in place of configuring admin groups.
func (a *AuthProvider) Reconcile(reqLogger logr.Logger, c client.Client, cluster *v1alpha1.VDICluster, adminPass string) error {
	return a.Setup(c, cluster)
}

// Close closes the underlying LDAP connection if it exists and then returns nil.
func (a *AuthProvider) Close() error {
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
		return nil
	}
	return nil
}
