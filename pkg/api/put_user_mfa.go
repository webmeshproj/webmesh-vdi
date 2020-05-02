package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/xlzd/gotp"
)

// swagger:operation PUT /api/users/{user}/mfa Users putUserMFARequest
// ---
// summary: Updates MFA configuration for the specified user.
// parameters:
// - name: user
//   in: path
//   description: The user to update
//   type: string
//   required: true
// - in: body
//   name: putUserMFARequest
//   description: The user details to update.
//   schema:
//     "$ref": "#/definitions/UpdateMFARequest"
// responses:
//   "200":
//     "$ref": "#/responses/updateMFAResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
//   "404":
//     "$ref": "#/responses/error"
func (d *desktopAPI) PutUserMFA(w http.ResponseWriter, r *http.Request) {
	username := apiutil.GetUserFromRequest(r)

	if _, err := d.auth.GetUser(username); err != nil {
		if errors.IsUserNotFoundError(err) {
			apiutil.ReturnAPINotFound(err, w)
			return
		}
		apiutil.ReturnAPIError(err, w)
		return
	}

	req := apiutil.GetRequestObject(r).(*v1alpha1.UpdateMFARequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	// We are enabling MFA
	if req.Enabled {
		newSecret := gotp.RandomSecret(32)
		if err := d.mfa.SetUserSecret(username, newSecret); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		apiutil.WriteJSON(&v1alpha1.UpdateMFAResponse{
			Enabled:         true,
			ProvisioningURI: gotp.NewDefaultTOTP(newSecret).ProvisioningUri(username, "kVDI"),
		}, w)
		return
	}

	// We are disabling MFA
	if err := d.mfa.DeleteUserSecret(username); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&v1alpha1.UpdateMFAResponse{
		Enabled: false,
	}, w)
}

// Request containing updates to a user
// swagger:parameters putUserMFARequest
type swaggerUpdateMFARequest struct {
	// in:body
	Body v1alpha1.UpdateMFARequest
}

// Session response
// swagger:response updateMFAResponse
type swaggerUpdateMFAResponse struct {
	// in:body
	Body v1alpha1.UpdateMFAResponse
}
