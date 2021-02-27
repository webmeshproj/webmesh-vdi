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
	"fmt"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/apis/meta/v1"
	rbacv1 "github.com/tinyzimmer/kvdi/apis/rbac/v1"

	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	ldapv3 "github.com/go-ldap/ldap/v3"
)

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *types.LoginRequest) (*types.AuthResult, error) {
	conn, err := a.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := a.bind(conn); err != nil {
		return nil, err
	}

	// fetch the role mappings
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}

	searchRequest := ldapv3.NewSearchRequest(
		a.getUserBase(),
		ldapv3.ScopeWholeSubtree, ldapv3.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(a.userFilter(), req.Username),
		a.userAttrs(),
		nil,
	)
	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) != 1 {
		return nil, errors.NewUserNotFoundError(fmt.Sprintf("Received %d matches for %s", len(sr.Entries), req.Username))
	}

	user := sr.Entries[0]

	if a.cluster.GetLDAPDoUserStatusCheck() {
		if strings.EqualFold(user.GetAttributeValue(a.cluster.GetLDAPUserStatusAttribute()), a.cluster.GetLDAPUserStatusDisabledValue()) {
			return nil, fmt.Errorf("User account %s is disabled", user.GetAttributeValue(a.cluster.GetLDAPUserIDAttribute()))
		}
	}

	// perform a bind to check the credentials
	if err := conn.Bind(user.DN, req.Password); err != nil {
		return nil, err
	}

	// make a new user object
	vdiUser := &types.VDIUser{
		Name:  req.Username,
		Roles: make([]*types.VDIUserRole, 0),
	}

	// we'll have to iterate our available roles and check if any have an annotation
	// binding it to one of this user's ldap groups
	boundRoles := make([]string, 0)
	userGroups := user.GetAttributeValues(a.cluster.GetLDAPUserGroupsAttribute())

	for _, role := range roles {
		boundRoles = appendRoleIfBound(boundRoles, userGroups, role)
	}

	vdiUser.Roles = apiutil.FilterUserRolesByNames(roles, boundRoles)

	// user is a regular user, check their ldap groups against any bound VDIRoles.
	return &types.AuthResult{User: vdiUser}, nil
}

func appendRoleIfBound(boundRoles, userGroups []string, role *rbacv1.VDIRole) []string {
	if annotations := role.GetAnnotations(); annotations != nil {
		if ldapGroups, ok := annotations[v1.LDAPGroupRoleAnnotation]; ok {
			boundGroups := strings.Split(ldapGroups, v1.AuthGroupSeparator)
			for _, group := range boundGroups {
				if group == "" {
					continue
				}
				if common.StringSliceContains(userGroups, group) {
					boundRoles = common.AppendStringIfMissing(boundRoles, role.GetName())
				}
			}
		}
	}
	return boundRoles
}
