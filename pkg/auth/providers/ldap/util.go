package ldap

import (
	"fmt"
)

func (a *AuthProvider) getUserBase() string {
	if base := a.cluster.GetLDAPSearchBase(); base != "" {
		return base
	}
	return a.baseDN
}

func (a *AuthProvider) userAttrs() []string {
	return []string{"cn", "dn", a.cluster.GetLDAPUserIDAttribute(), a.cluster.GetLDAPUserGroupsAttribute()}
}

func (a *AuthProvider) userFilter() string {
	return fmt.Sprintf("(%s=%%s)", a.cluster.GetLDAPUserIDAttribute())
}

func (a *AuthProvider) groupUsersFilter() string {
	return fmt.Sprintf("(%s=%%s)", a.cluster.GetLDAPUserGroupsAttribute())
}
