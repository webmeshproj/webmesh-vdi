package local

import (
	"errors"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	vdierrors "github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// GetUsers implements AuthProvider and serves a GET /api/users request
func (a *LocalAuthProvider) GetUsers(w http.ResponseWriter, r *http.Request) {
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	users, err := a.listUsers()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	res := make([]*v1alpha1.VDIUser, 0)
	for _, user := range users {
		res = append(res, &v1alpha1.VDIUser{
			Name:  user.Username,
			Roles: apiutil.FilterUserRolesByNames(roles, user.Groups),
		})
	}

	apiutil.WriteJSON(res, w)
}

// PostUsers implements AuthProvider and serves a POST /api/users request
func (a *LocalAuthProvider) PostUsers(w http.ResponseWriter, r *http.Request) {
	req := apiutil.GetRequestObject(r).(*v1alpha1.CreateUserRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	passwdHash, err := common.HashPassword(req.Password)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	user := &LocalUser{
		Username:     req.Username,
		PasswordHash: passwdHash,
		Groups:       req.Roles,
	}
	if err := a.createUser(user); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// GetUser implements AuthProvider and serves a GET /api/users/{user} request
func (a *LocalAuthProvider) GetUser(w http.ResponseWriter, r *http.Request) {
	userName := apiutil.GetUserFromRequest(r)
	user, err := a.getUser(userName)
	if err != nil {
		if vdierrors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
	}

	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&v1alpha1.VDIUser{
		Name:  user.Username,
		Roles: apiutil.FilterUserRolesByNames(roles, user.Groups),
	}, w)
}

// PutUser implements AuthProvider and serves a PUT /api/users/{user} request
func (a *LocalAuthProvider) PutUser(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)
	req := apiutil.GetRequestObject(r).(*v1alpha1.UpdateUserRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	user := &LocalUser{Username: username}
	if len(req.Roles) != 0 {
		user.Groups = req.Roles
	}
	if req.Password != "" {
		passwdHash, err := common.HashPassword(req.Password)
		if err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		user.PasswordHash = passwdHash
	}
	if err := a.updateUser(user); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// DeleteUser implements AuthProvider and serves a DELETE /api/users/{user} request
func (a *LocalAuthProvider) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)
	if err := a.deleteUser(username); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}
