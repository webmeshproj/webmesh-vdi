package local

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

// GetUsers implements AuthProvider and serves a GET /api/users request
func (a *LocalAuthProvider) GetUsers() ([]*v1alpha1.VDIUser, error) {
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}
	users, err := a.listUsers()
	if err != nil {
		return nil, err
	}
	res := make([]*v1alpha1.VDIUser, 0)
	for _, user := range users {
		res = append(res, &v1alpha1.VDIUser{
			Name:  user.Username,
			Roles: apiutil.FilterUserRolesByNames(roles, user.Groups),
		})
	}

	return res, nil
}

// CreateUser implements AuthProvider and serves a POST /api/users request
func (a *LocalAuthProvider) CreateUser(req *v1alpha1.CreateUserRequest) error {
	passwdHash, err := common.HashPassword(req.Password)
	if err != nil {
		return err
	}
	user := &LocalUser{
		Username:     req.Username,
		PasswordHash: passwdHash,
		Groups:       req.Roles,
	}
	return a.createUser(user)
}

// GetUser implements AuthProvider and serves a GET /api/users/{user} request
func (a *LocalAuthProvider) GetUser(username string) (*v1alpha1.VDIUser, error) {
	user, err := a.getUser(username)
	if err != nil {
		return nil, err
	}

	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.VDIUser{
		Name:  user.Username,
		Roles: apiutil.FilterUserRolesByNames(roles, user.Groups),
	}, nil
}

// UpdateUser implements AuthProvider and serves a PUT /api/users/{user} request
func (a *LocalAuthProvider) UpdateUser(username string, req *v1alpha1.UpdateUserRequest) error {
	user := &LocalUser{Username: username}
	if len(req.Roles) != 0 {
		user.Groups = req.Roles
	}
	if req.Password != "" {
		passwdHash, err := common.HashPassword(req.Password)
		if err != nil {
			return err
		}
		user.PasswordHash = passwdHash
	}
	return a.updateUser(user)
}

// DeleteUser implements AuthProvider and serves a DELETE /api/users/{user} request
func (a *LocalAuthProvider) DeleteUser(username string) error {
	return a.deleteUser(username)
}
