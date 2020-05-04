package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/xlzd/gotp"
)

// swagger:route POST /api/authorize Auth authorizeRequest
// Authorizes a JWT token with a one time password.
// responses:
//   200: sessionResponse
//   400: error
//   403: error
func (d *desktopAPI) PostAuthorize(w http.ResponseWriter, r *http.Request) {
	userSession := apiutil.GetRequestUserSession(r)

	secret, err := d.mfa.GetUserSecret(userSession.User.Name)
	if err != nil {
		if !errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPIError(err, w)
			return
		}
		// The user does not require MFA - this shouldn't happen but go ahead
		// and send back an authorized token
		d.returnNewJWT(w, userSession.User, true)
		return
	}

	// retrieve the OTP from the request
	req := apiutil.GetRequestObject(r).(*v1alpha1.AuthorizeRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	totp := gotp.NewDefaultTOTP(secret)

	if totp.Now() != req.OTP {
		apiutil.ReturnAPIForbidden(nil, "Invalid MFA Code", w)
		return
	}

	d.returnNewJWT(w, userSession.User, true)
}

// Request containing a one-time password.
// swagger:parameters authorizeRequest
type swaggerAuthorizeRequest struct {
	// in:body
	Body v1alpha1.AuthorizeRequest
}
