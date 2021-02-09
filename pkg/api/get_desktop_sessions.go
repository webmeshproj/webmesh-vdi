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
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// swagger:route GET /api/sessions Sessions getDesktopSessions
// Retrieves a list of currently active desktop sessions and their status.
// responses:
//   200: desktopSessionsResponse
//   400: error
//   403: error
func (d *desktopAPI) GetDesktopSessions(w http.ResponseWriter, r *http.Request) {
	desktops := &v1alpha1.DesktopList{}
	displayLocks := &corev1.ConfigMapList{}
	audioLocks := &corev1.ConfigMapList{}

	// retrieve all desktops for this cluster
	if err := d.client.List(context.TODO(), desktops, client.InNamespace(metav1.NamespaceAll), d.vdiCluster.GetClusterDesktopsSelector()); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// retrieve all active display locks
	if err := d.client.List(
		context.TODO(),
		displayLocks,
		client.InNamespace(metav1.NamespaceAll),
		client.MatchingLabels(d.vdiCluster.GetComponentLabels("display-lock")),
	); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// retrieve all active audio locks
	if err := d.client.List(
		context.TODO(),
		audioLocks,
		client.InNamespace(metav1.NamespaceAll),
		client.MatchingLabels(d.vdiCluster.GetComponentLabels("audio-lock")),
	); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// initialize a response
	res := &v1.DesktopSessionsResponse{
		Sessions: make([]*v1.DesktopSession, 0),
	}

	// iterate desktops and parse properties and connection status
	for _, desktop := range desktops.Items {
		sess := &v1.DesktopSession{
			Name:           desktop.GetName(),
			Namespace:      desktop.GetNamespace(),
			User:           desktop.GetUser(),
			ServiceAccount: desktop.GetServiceAccount(),
			Status:         getSessionStatus(d.vdiCluster, desktop, displayLocks.Items, audioLocks.Items),
		}
		res.Sessions = append(res.Sessions, sess)
	}

	// return the response
	apiutil.WriteJSON(res, w)
}

// getSessionStatus iterates the current locks and builds a session object for the given desktop.
// TODO: Getters for the names of locks, sprintf calls also present in get_websockify.go. This function
// could also be optimized to pop found locks off for future iterations.
func getSessionStatus(cluster *v1alpha1.VDICluster, desktop v1alpha1.Desktop, displayLocks, audioLocks []corev1.ConfigMap) *v1.DesktopSessionStatus {
	status := &v1.DesktopSessionStatus{
		Display: &v1.ConnectionStatus{Connected: false},
		Audio:   &v1.ConnectionStatus{Connected: false},
	}
	displayLockName := fmt.Sprintf("display-%s-%s", desktop.GetNamespace(), desktop.GetName())
	audioLockName := fmt.Sprintf("audio-%s-%s", desktop.GetNamespace(), desktop.GetName())

	// iterate display locks and populate the status if one matches this desktop
	for _, lock := range displayLocks {
		if lock.GetName() == displayLockName {
			status.Display.Connected = true
			if len(lock.OwnerReferences) == 0 {
				status.Display.ProxyPod = "<unknown>"
			} else {
				status.Display.ProxyPod = fmt.Sprintf("%s/%s", cluster.GetCoreNamespace(), lock.OwnerReferences[0].Name)
			}
			if lock.GetLabels() != nil {
				if clientAddr, ok := lock.Labels[v1.ClientAddrLabel]; ok {
					status.Display.ClientAddr = clientAddr

				} else {
					status.Display.ClientAddr = "<unknown>"
				}
			} else {
				status.Display.ClientAddr = "<unknown>"
			}
		}
	}

	// iterate audio locks and populate the status if one matches this desktop
	for _, lock := range audioLocks {
		if lock.GetName() == audioLockName {
			status.Audio.Connected = true
			if len(lock.OwnerReferences) == 0 {
				status.Audio.ProxyPod = "<unknown>"
			} else {
				status.Audio.ProxyPod = fmt.Sprintf("%s/%s", cluster.GetCoreNamespace(), lock.OwnerReferences[0].Name)
			}
			if lock.GetLabels() != nil {
				if clientAddr, ok := lock.Labels[v1.ClientAddrLabel]; ok {
					status.Audio.ClientAddr = clientAddr

				} else {
					status.Audio.ClientAddr = "<unknown>"
				}
			} else {
				status.Audio.ClientAddr = "<unknown>"
			}
		}
	}

	return status
}

// Desktop Sessions Response
// swagger:response desktopSessionsResponse
type swaggerDesktopSessionsResponse struct {
	// in:body
	Body v1.DesktopSessionsResponse
}
