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
	"github.com/tinyzimmer/kvdi/pkg/types"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	rbacutil "github.com/tinyzimmer/kvdi/pkg/util/rbac"

	ldapv3 "github.com/go-ldap/ldap/v3"
)

// GetUsers should return a list of VDIUsers.
func (a *AuthProvider) GetUsers() ([]*types.VDIUser, error) {
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

	vdiUsers := make([]*types.VDIUser, 0)
	for _, role := range roles {

		userRole := rbacutil.VDIRoleToUserRole(&role)

		if annotations := role.GetAnnotations(); annotations != nil {
			if ldapGroupStr, ok := annotations[v1.LDAPGroupRoleAnnotation]; ok {
				groups := strings.Split(ldapGroupStr, v1.AuthGroupSeparator)
			GroupLoop:
				for _, group := range groups {
					if group == "" {
						continue GroupLoop
					}
					searchRequest := ldapv3.NewSearchRequest(
						a.getUserBase(),
						ldapv3.ScopeWholeSubtree, ldapv3.NeverDerefAliases, 0, 0, false,
						fmt.Sprintf(a.groupUsersFilter(), group),
						a.userAttrs(),
						nil,
					)
					sr, err := conn.Search(searchRequest)
					if err != nil {
						return nil, err
					}
					for _, entry := range sr.Entries {
						vdiUsers = appendUser(vdiUsers, entry.GetAttributeValue(a.cluster.GetLDAPUserIDAttribute()), userRole)
					}
				}
			}
		}
	}

	return vdiUsers, nil

}

// GetUser should retrieve a single VDIUser.
func (a *AuthProvider) GetUser(username string) (*types.VDIUser, error) {
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
		fmt.Sprintf(a.userFilter(), username),
		a.userAttrs(),
		nil,
	)
	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) != 1 {
		return nil, errors.NewUserNotFoundError(fmt.Sprintf("Received %d matches for %s", len(sr.Entries), username))
	}

	user := sr.Entries[0]

	vdiUser := &types.VDIUser{
		Name:  username,
		Roles: make([]*types.VDIUserRole, 0),
	}

RoleLoop:
	for _, role := range roles {
		if annotations := role.GetAnnotations(); annotations != nil {
			if ldapGroupStr, ok := annotations[v1.LDAPGroupRoleAnnotation]; ok {
			GroupLoop:
				for _, group := range strings.Split(ldapGroupStr, v1.AuthGroupSeparator) {
					if group == "" {
						continue GroupLoop
					}
					if common.StringSliceContains(user.GetAttributeValues(a.cluster.GetLDAPUserGroupsAttribute()), group) {
						vdiUser.Roles = append(vdiUser.Roles, rbacutil.VDIRoleToUserRole(&role))
						continue RoleLoop
					}
				}
			}
		}
	}

	return vdiUser, nil
}

// CreateUser should handle any logic required to register a new user in kVDI.
func (a *AuthProvider) CreateUser(*types.CreateUserRequest) error {
	return errors.New("Creating users is not supported when using LDAP authentication")
}

// UpdateUser should update a VDIUser.
func (a *AuthProvider) UpdateUser(string, *types.UpdateUserRequest) error {
	return errors.New("Updating users is not supported when using LDAP authentication")
}

// DeleteUser should remove a VDIUser.
func (a *AuthProvider) DeleteUser(string) error {
	return errors.New("Deleting users is not supported when using LDAP authentication")
}

func appendUser(vdiUsers []*types.VDIUser, name string, role *types.VDIUserRole) []*types.VDIUser {
	for _, user := range vdiUsers {
		if user.Name == name {
			for _, userRole := range user.Roles {
				if userRole.Name == role.Name {
					return vdiUsers
				}
			}
			user.Roles = append(user.Roles, role)
			return vdiUsers
		}
	}
	return append(vdiUsers, &types.VDIUser{
		Name:  name,
		Roles: []*types.VDIUserRole{role},
	})
}
