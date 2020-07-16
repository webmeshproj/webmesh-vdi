package ldap

import (
	"fmt"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	ldapv3 "github.com/go-ldap/ldap/v3"
)

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *v1alpha1.LoginRequest) (*v1alpha1.AuthResult, error) {
	a.mux.Lock()
	defer a.mux.Unlock()
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
	sr, err := a.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) != 1 {
		return nil, errors.NewUserNotFoundError(fmt.Sprintf("Received %d matches for %s", len(sr.Entries), req.Username))
	}

	user := sr.Entries[0]

	// perform a bind to check the credentials
	if err := a.conn.Bind(user.DN, req.Password); err != nil {
		// rebind to svc account first
		if rbErr := a.bind(); rbErr != nil {
			ldapLog.Error(rbErr, "Failed to rebind to service account")
		}
		return nil, err
	}

	// rebind to svc account first
	if err = a.bind(); err != nil {
		return nil, err
	}

	// make a new user object
	vdiUser := &v1alpha1.VDIUser{
		Name:  req.Username,
		Roles: make([]*v1alpha1.VDIUserRole, 0),
	}

	// we'll have to iterate our available roles and check if any have an annotation
	// binding it to one of this user's ldap groups
	boundRoles := make([]string, 0)
RoleLoop:
	for _, role := range roles {
		if annotations := role.GetAnnotations(); annotations != nil {
			if ldapGroups, ok := annotations[v1alpha1.LDAPGroupRoleAnnotation]; ok {
				boundGroups := strings.Split(ldapGroups, ";")
				for _, group := range boundGroups {
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
	return &v1alpha1.AuthResult{User: vdiUser}, nil
}
