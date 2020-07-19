package ldap

import (
	"fmt"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	ldapv3 "github.com/go-ldap/ldap/v3"
)

// GetUsers should return a list of VDIUsers.
func (a *AuthProvider) GetUsers() ([]*v1.VDIUser, error) {
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

	vdiUsers := make([]*v1.VDIUser, 0)
	for _, role := range roles {

		userRole := role.ToUserRole()

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
						fmt.Sprintf(groupUsersFilter, group),
						userAttrs,
						nil,
					)
					sr, err := conn.Search(searchRequest)
					if err != nil {
						return nil, err
					}
					for _, entry := range sr.Entries {
						vdiUsers = appendUser(vdiUsers, entry.GetAttributeValue("uid"), userRole)
					}
				}
			}
		}
	}

	return vdiUsers, nil

}

// GetUser should retrieve a single VDIUser.
func (a *AuthProvider) GetUser(username string) (*v1.VDIUser, error) {
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
		fmt.Sprintf(userFilter, username),
		userAttrs,
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

	vdiUser := &v1.VDIUser{
		Name:  username,
		Roles: make([]*v1.VDIUserRole, 0),
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
					if common.StringSliceContains(user.GetAttributeValues("memberOf"), group) {
						vdiUser.Roles = append(vdiUser.Roles, role.ToUserRole())
						continue RoleLoop
					}
				}
			}
		}
	}

	return vdiUser, nil
}

// CreateUser should handle any logic required to register a new user in kVDI.
func (a *AuthProvider) CreateUser(*v1.CreateUserRequest) error {
	return errors.New("Creating users is not supported when using LDAP authentication")
}

// UpdateUser should update a VDIUser.
func (a *AuthProvider) UpdateUser(string, *v1.UpdateUserRequest) error {
	return errors.New("Updating users is not supported when using LDAP authentication")
}

// DeleteUser should remove a VDIUser.
func (a *AuthProvider) DeleteUser(string) error {
	return errors.New("Deleting users is not supported when using LDAP authentication")
}

func appendUser(vdiUsers []*v1.VDIUser, name string, role *v1.VDIUserRole) []*v1.VDIUser {
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
	return append(vdiUsers, &v1.VDIUser{
		Name:  name,
		Roles: []*v1.VDIUserRole{role},
	})
}
