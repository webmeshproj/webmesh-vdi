/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

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
