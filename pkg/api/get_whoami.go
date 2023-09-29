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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// swagger:route GET /api/whoami Miscellaneous whoAmI
// Retrieves information about the current user session.
// responses:
//
//	200: userResponse
//	403: error
//	500: error
func (d *desktopAPI) GetWhoAmI(w http.ResponseWriter, r *http.Request) {
	// retrieve the user session from the request
	session := apiutil.GetRequestUserSession(r)
	// retrieve all desktops for this user and populate the Sessions field
	desktops := &desktopsv1.SessionList{}
	if err := d.client.List(context.TODO(), desktops, client.InNamespace(metav1.NamespaceAll), d.vdiCluster.GetUserDesktopsSelector(session.User.Name)); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	// If the user has any active desktops, append them to the response
	if len(desktops.Items) > 0 {
		session.User.Sessions = make([]*types.DesktopSession, len(desktops.Items))
		for idx, desktop := range desktops.Items {
			session.User.Sessions[idx] = &types.DesktopSession{
				Name:      desktop.GetName(),
				Namespace: desktop.GetNamespace(),
				User:      desktop.GetUser(),
				Template:  desktop.GetTemplateName(),
			}
		}
	}
	apiutil.WriteJSON(session.User, w)
}
