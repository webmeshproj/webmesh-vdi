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
	"context"
	"net/http"

	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
)

func allowSameUser(d *desktopAPI, reqUser *types.VDIUser, r *http.Request) (allowed, owner bool, err error) {
	pathUser := apiutil.GetUserFromRequest(r)
	if reqUser.Name != pathUser {
		return false, false, nil
	}
	// make sure the user isn't trying to change their permission level
	allowed, _, err = denyUserElevatePerms(d, reqUser, r)
	return allowed, true, err
}

func allowSessionOwner(d *desktopAPI, reqUser *types.VDIUser, r *http.Request) (allowed, owner bool, err error) {
	nn := apiutil.GetNamespacedNameFromRequest(r)
	found := &desktopsv1.Session{}
	if err := d.client.Get(context.TODO(), nn, found); err != nil {
		return false, false, err
	}
	userDesktopLabels := d.vdiCluster.GetUserDesktopSelector(reqUser.Name)
	// extra safety check - cant accurately determine ownership without labels
	if found.GetLabels() == nil {
		return false, false, nil
	}
	for key, val := range userDesktopLabels {
		if expected, ok := found.GetLabels()[key]; !ok {
			return false, false, nil
		} else if expected != val {
			return false, false, nil
		}
	}
	return true, true, nil
}

func allowAll(d *desktopAPI, reqUser *types.VDIUser, r *http.Request) (allowed, owner bool, err error) {
	return true, false, nil
}
