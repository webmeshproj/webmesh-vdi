package api

import (
	"net/http"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

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

	// Only verify user if not using OIDC. We don't have a way to verify the user
	// otherwise. This does leave the door open for someone with access to this endpoint
	// to go rogue and flood the secrets with bad users.
	if !d.vdiCluster.IsUsingOIDCAuth() {
		if _, err := d.auth.GetUser(username); err != nil {
			if errors.IsUserNotFoundError(err) {
				apiutil.ReturnAPINotFound(err, w)
				return
			}
			apiutil.ReturnAPIError(err, w)
			return
		}
	}

	req := apiutil.GetRequestObject(r).(*v1.UpdateMFARequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	// We are enabling MFA
	if req.Enabled {
		// https://github.com/xlzd/gotp/blob/master/utils.go#L79
		//Only uses uppercase characters and digits
		newSecret := gotp.RandomSecret(32)
		if err := d.mfa.SetUserMFAStatus(username, newSecret, false); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		apiutil.WriteJSON(&v1.MFAResponse{
			Enabled:         true,
			Verified:        false,
			ProvisioningURI: gotp.NewDefaultTOTP(newSecret).ProvisioningUri(username, "kVDI"),
		}, w)
		return
	}

	// We are disabling MFA
	if err := d.mfa.DeleteUserSecret(username); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&v1.MFAResponse{
		Enabled: false,
	}, w)
}

// Request containing updates to a user
// swagger:parameters putUserMFARequest
type swaggerUpdateMFARequest struct {
	// in:body
	Body v1.UpdateMFARequest
}

// Session response
// swagger:response updateMFAResponse
type swaggerUpdateMFAResponse struct {
	// in:body
	Body v1.MFAResponse
}
