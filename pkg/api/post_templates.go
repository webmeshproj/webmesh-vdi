package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:route POST /api/templates Templates postTemplateRequest
// Create a new DesktopTemplate in kVDI.
// responses:
//   200: boolResponse
//   400: error
//   403: error
func (d *desktopAPI) PostDesktopTemplates(w http.ResponseWriter, r *http.Request) {
	tmpl := apiutil.GetRequestObject(r).(*v1alpha1.DesktopTemplate)
	if tmpl == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}
	if err := d.client.Create(context.TODO(), tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

// Request containing a new user
// swagger:parameters postTemplateRequest
type swaggerCreateTemplateRequest struct {
	// in:body
	Body v1alpha1.DesktopTemplate
}
