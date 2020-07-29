package api

import (
	"net/http"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

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

	// If this is a GET request, we are at the second-phase of a redirect auth-flow.
	if r.Method == http.MethodGet {
		// Create a login request to pass to the auth backend containing just the
		// raw request object. The backend provider should know how to use it to
		// return valid claims.
		req := &v1.LoginRequest{}
		req.SetRequest(r)

		// pass the request object to the auth backend, it should know how to handle a
		// GET separately. The backend needs to generate claims that it can then
		// provide on a subsequent POST with the initial state token.
		_, err := d.auth.Authenticate(req)
		if err != nil {
			apiLogger.Error(err, "Failure handling auth callback")
			apiutil.ReturnAPIError(err, w)
			return
		}

		// redirect back to home page. the ui knows to use it's existing state token
		// and attempt anonymous login. The next POST should return the proper claims.
		http.Redirect(w, r, "/#/login", http.StatusFound)
		return
	}

	// Retrieve the request object
	req := apiutil.GetRequestObject(r).(*v1.LoginRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	// Not needed at the moment, but in case further use of the request object
	// is needed in the authentication flow.
	req.SetRequest(r)

	// Pass the request to the provider
	result, err := d.auth.Authenticate(req)
	if err != nil {
		apiLogger.Error(err, "Authentication failed, checking if anonymous is allowed")
		// Allow anonymous if set in the configuration
		if req.GetUsername() == userAnonymous && d.vdiCluster.AnonymousAllowed() {
			result := &v1.AuthResult{
				User: &v1.VDIUser{
					Name:  userAnonymous,
					Roles: []*v1.VDIUserRole{d.vdiCluster.GetLaunchTemplatesRole().ToUserRole()},
				},
			}
			d.returnNewJWT(w, result, true, req.GetState())
			return
		}
		// If it's not an actual credential error, it will still be logged server side,
		// but always tell the user 'Invalid credentials'.
		apiutil.ReturnAPIForbidden(err, "Invalid credentials", w)
		return
	}

	// Check if the auth provider requires a redirect
	if result.RedirectURL != "" {
		w.Header().Set("X-Redirect", result.RedirectURL)
		apiutil.WriteJSON(map[string]string{
			"message": "Authentication requires sign-in to an external resource",
			"state":   req.GetState(),
		}, w)
		return
	}

	d.checkMFAAndReturnJWT(w, result, req.GetState())
}

func (d *desktopAPI) checkMFAAndReturnJWT(w http.ResponseWriter, result *v1.AuthResult, state string) {
	// check if MFA is configured for the user and that they have verified their secret
	if _, verified, err := d.mfa.GetUserMFAStatus(result.User.Name); err != nil || !verified {
		// Return any error that isn't a not found error
		if err != nil && !errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIError(err, w)
			return
		}
		// The user does not require MFA
		d.returnNewJWT(w, result, true, state)
		return
	}

	// the user requires MFA
	d.returnNewJWT(w, result, false, state)
}

// Login request
// swagger:parameters loginRequest
type swaggerLoginRequest struct {
	// in:body
	Body v1.LoginRequest
}
