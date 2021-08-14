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
	"github.com/kvdi/kvdi/pkg/auth/common"
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
