package api

import (
	"errors"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

const userAnonymous = "anonymous"

// PostLogin handles a login request. The request object is passed to the
// authentication provider, and then a token is created and returned to the user
// based off the data returned by the provider.
func (d *desktopAPI) PostLogin(w http.ResponseWriter, r *http.Request) {
	req := apiutil.GetRequestObject(r).(*v1alpha1.LoginRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	// Allow anonymous if set in the configuration
	if req.Username == userAnonymous && d.vdiCluster.AnonymousAllowed() {
		user := &v1alpha1.VDIUser{
			Name:  userAnonymous,
			Roles: []*v1alpha1.VDIUserRole{d.vdiCluster.GetLaunchTemplatesRole().ToUserRole()},
		}
		d.returnNewJWT(w, user, true)
		return
	}

	// Pass the request to the provider, any error is a failure.
	result, err := d.auth.Authenticate(req)
	if err != nil {
		// If it's not an actual credential error, it will still be logged server side,
		// but always tell the user 'Invalid credentials'.
		apiutil.ReturnAPIForbidden(err, "Invalid credentials", w)
		return
	}

	d.returnNewJWT(w, result.User, true)
}

func (d *desktopAPI) returnNewJWT(w http.ResponseWriter, user *v1alpha1.VDIUser, authorized bool) {
	// fetch the JWT signing secret
	secret, err := d.secrets.ReadSecret(v1alpha1.JWTSecretKey, true)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// create a new token
	claims, newToken, err := apiutil.GenerateJWT(secret, user, authorized)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// return the token to the user
	apiutil.WriteJSON(&v1alpha1.SessionResponse{
		Token:     newToken,
		ExpiresAt: claims.ExpiresAt,
		User:      user,
	}, w)
}
