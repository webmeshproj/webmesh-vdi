package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

func (d *desktopAPI) Healthz(w http.ResponseWriter, r *http.Request) {}

func (d *desktopAPI) Readyz(w http.ResponseWriter, r *http.Request) {
	if errs := d.checkReadiness(); len(errs) != 0 {
		apiutil.ReturnAPIErrors(errs, w)
		return
	}
	apiutil.WriteOK(w)
}

func (d *desktopAPI) checkReadiness() []error {
	errs := make([]error, 0)
	if d.auth == nil {
		errs = append(errs, errors.New("Authentication has not been setup yet"))
	}
	if d.secrets == nil {
		errs = append(errs, errors.New("Secrets storage has not been setup yet"))
	}
	if d.mfa == nil {
		errs = append(errs, errors.New("MFA storage has not been setup yet "))
	}
	return errs
}
