package ldap

func (a *AuthProvider) getUserBase() string {
	if base := a.cluster.GetLDAPSearchBase(); base != "" {
		return base
	}
	return a.baseDN
}
