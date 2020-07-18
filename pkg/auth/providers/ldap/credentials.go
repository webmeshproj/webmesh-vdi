package ldap

import (
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
)

// getCredentials returns the bind credentials for the configured service account.
func (a *AuthProvider) getCredentials() (user, passw string, err error) {

	userKey := a.cluster.GetLDAPUserDNKey()
	passKey := a.cluster.GetLDAPPasswordKey()

	secrets, err := common.GetAuthSecrets(a.client, a.cluster, a.secrets, userKey, passKey)
	if err != nil {
		return "", "", err
	}
	return secrets[userKey], secrets[passKey], nil
}
