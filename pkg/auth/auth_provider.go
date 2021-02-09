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

// Package auth contains the methods for retrieving the AuthProvider for a
// given cluster.
package auth

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/common"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/ldap"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/local"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/oidc"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
)

// GetAuthProvider returns the authentication provider for the given VDICluster. The secret engine passed
// to the provider is assumed to already be setup.
func GetAuthProvider(cluster *v1alpha1.VDICluster, s *secrets.SecretEngine) common.AuthProvider {
	if cluster.IsUsingLDAPAuth() {
		return ldap.New(s)
	}
	if cluster.IsUsingOIDCAuth() {
		return oidc.New(s)
	}
	return local.New(s)
}
