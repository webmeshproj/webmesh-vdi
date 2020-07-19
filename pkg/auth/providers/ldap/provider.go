package ldap

import (
	"crypto/tls"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/secrets"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const userFilter = "(uid=%s)"
const groupUsersFilter = "(memberOf=%s)"

var userAttrs = []string{"cn", "dn", "uid", "memberOf", "accountStatus"}

// AuthProvider implements an auth provider that uses an LDAP server as the
// authentication backend. Access to groups in LDAP is supplied through annotations
// on VDIRoles.
type AuthProvider struct {
	common.AuthProvider

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
	// the base DN for the connected LDAP server
	baseDN string
}

// Blank assignment to make sure AuthProvider satisfies the interface.
var _ common.AuthProvider = &AuthProvider{}

// New returns a new LDAPAuthProvider.
func New() common.AuthProvider {
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
func (a *AuthProvider) Reconcile(reqLogger logr.Logger, c client.Client, cluster *v1alpha1.VDICluster, adminPass string) error {
	return a.Setup(c, cluster)
}

// Close just returns nil as connections are not persistent
func (a *AuthProvider) Close() error {
	return nil
}
