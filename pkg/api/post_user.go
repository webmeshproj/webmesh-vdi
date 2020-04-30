package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
)

// swagger:route POST /api/users Users postUserRequest
// Create a new user in kVDI.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) CreateUser(w http.ResponseWriter, r *http.Request) {}

// Implemented by the auth provider

// Request containing a new user
// swagger:parameters postUserRequest
type swaggerCreateUserRequest struct {
	// in:body
	Body v1alpha1.CreateUserRequest
}
