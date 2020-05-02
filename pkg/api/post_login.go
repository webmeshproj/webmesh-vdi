package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

const userAnonymous = "anonymous"

// swagger:route POST /api/login Auth loginRequest
// Retrieves a new JWT token. This route may behave differently depending on the auth provider.
// responses:
//   200: sessionResponse
//   400: error
//   403: error
//   500: error
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

	// check if MFA is configured for the user
	if _, err := d.mfa.GetUserSecret(result.User.Name); err != nil {
		if !errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIError(err, w)
			return
		}
		// The user does not require MFA
		d.returnNewJWT(w, result.User, true)
		return
	}

	// the user requires MFA
	d.returnNewJWT(w, result.User, false)
}

// Login request
// swagger:parameters loginRequest
type swaggerLoginRequest struct {
	// in:body
	Body v1alpha1.LoginRequest
}
