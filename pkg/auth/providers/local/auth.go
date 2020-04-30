package local

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// Authenticate implements AuthProvider and simply checks the provided password
// in the request against the hash in the file.
func (a *LocalAuthProvider) Authenticate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	req := &v1alpha1.LoginRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	user := &v1alpha1.VDIUser{
		Name:  req.Username,
		Roles: make([]*v1alpha1.VDIUserRole, 0),
	}

	localUser, err := a.getUser(req.Username)
	if err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIForbidden(nil, "Invalid credentials", w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	if !localUser.PasswordMatchesHash(req.Password) {
		apiutil.ReturnAPIForbidden(nil, "Invalid credentials", w)
		return
	}

	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	user.Roles = apiutil.FilterUserRolesByNames(roles, localUser.Groups)

	secret, err := apiutil.GetJWTSecret()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	claims, newToken, err := apiutil.GenerateJWT(secret, user)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	response := &v1alpha1.SessionResponse{
		Token:     newToken,
		ExpiresAt: claims.ExpiresAt,
		User:      user,
	}

	apiutil.WriteJSON(response, w)
}
