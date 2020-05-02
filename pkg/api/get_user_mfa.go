package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/xlzd/gotp"
)

// swagger:operation GET /api/users/{user}/mfa Users getUserMFARequest
// ---
// summary: Retrieves MFA status for the given user.
// parameters:
// - name: user
//   in: path
//   description: The user to query
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/getMFAResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetUserMFA(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)

	if _, err := d.auth.GetUser(username); err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	secret, err := d.mfa.GetUserSecret(username)
	if err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.WriteJSON(&v1alpha1.UpdateMFAResponse{
				Enabled: false,
			}, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&v1alpha1.UpdateMFAResponse{
		Enabled:         true,
		ProvisioningURI: gotp.NewDefaultTOTP(secret).ProvisioningUri(username, "kVDI"),
	}, w)
}

// Session response
// swagger:response getMFAResponse
type swaggerGetMFAResponse struct {
	// in:body
	Body v1alpha1.UpdateMFAResponse
}
