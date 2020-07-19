package ldap

import (
	"fmt"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	ldapv3 "github.com/go-ldap/ldap/v3"
)

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *v1.LoginRequest) (*v1.AuthResult, error) {
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
		fmt.Sprintf(userFilter, req.Username),
		userAttrs,
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

	if strings.ToLower(user.GetAttributeValue("accountStatus")) != "active" {
		return nil, fmt.Errorf("User account %s is disabled", user.GetAttributeValue("uid"))
	}

	// perform a bind to check the credentials
	if err := conn.Bind(user.DN, req.Password); err != nil {
		return nil, err
	}

	// make a new user object
	vdiUser := &v1.VDIUser{
		Name:  req.Username,
		Roles: make([]*v1.VDIUserRole, 0),
	}

	// we'll have to iterate our available roles and check if any have an annotation
	// binding it to one of this user's ldap groups
	boundRoles := make([]string, 0)
RoleLoop:
	for _, role := range roles {
		if annotations := role.GetAnnotations(); annotations != nil {
			if ldapGroups, ok := annotations[v1.LDAPGroupRoleAnnotation]; ok {
				boundGroups := strings.Split(ldapGroups, v1.AuthGroupSeparator)
			GroupLoop:
				for _, group := range boundGroups {
					if group == "" {
						continue GroupLoop
					}
					if common.StringSliceContains(user.GetAttributeValues("memberOf"), group) {
						boundRoles = common.AppendStringIfMissing(boundRoles, role.GetName())
						continue RoleLoop
					}
				}
			}
		}
	}

	vdiUser.Roles = apiutil.FilterUserRolesByNames(roles, boundRoles)

	// user is a regular user, check their ldap groups against any bound VDIRoles.
	return &v1.AuthResult{User: vdiUser}, nil
}
