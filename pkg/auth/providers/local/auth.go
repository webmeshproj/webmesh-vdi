package local

import (
	"errors"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// Authenticate implements AuthProvider and simply checks the provided password
// in the request against the hash in the file.
func (a *AuthProvider) Authenticate(req *v1.LoginRequest) (*v1.AuthResult, error) {

	user := &v1.VDIUser{
		Name:  req.Username,
		Roles: make([]*v1.VDIUserRole, 0),
	}

	localUser, err := a.getUser(req.Username)
	if err != nil {
		return nil, err
	}

	if !localUser.PasswordMatchesHash(req.Password) {
		return nil, errors.New("Invalid credentials")
	}

	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}

	user.Roles = apiutil.FilterUserRolesByNames(roles, localUser.Groups)
	return &v1.AuthResult{User: user}, nil
}
