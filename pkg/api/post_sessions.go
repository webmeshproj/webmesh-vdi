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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Request for a new desktop session
// swagger:parameters postSessionRequest
type swaggerCreateSessionRequest struct {
	// in:body
	Body types.CreateSessionRequest
}

// New session response
// swagger:response postSessionResponse
type swaggerCreateSessionResponse struct {
	// in:body
	Body types.CreateSessionResponse
}

// swagger:route POST /api/sessions Sessions postSessionRequest
// Creates a new desktop session with the given parameters.
// responses:
//
//	200: postSessionResponse
//	400: error
//	403: error
func (d *desktopAPI) StartDesktopSession(w http.ResponseWriter, r *http.Request) {
	sess := apiutil.GetRequestUserSession(r)
	req := apiutil.GetRequestObject(r).(*types.CreateSessionRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	if max := d.vdiCluster.GetMaxSessionsPerUser(); max > 0 {
		desktops := &desktopsv1.SessionList{}
		if err := d.client.List(context.TODO(), desktops, client.InNamespace(metav1.NamespaceAll), client.MatchingLabels(d.vdiCluster.GetUserDesktopSelector(sess.User.Name))); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		if len(desktops.Items) >= max {
			apiutil.ReturnAPIError(fmt.Errorf("%s has reached the maximum allowed (%d) running desktops", sess.User.Name, max), w)
			return
		}
	}

	tmplnn := ktypes.NamespacedName{Name: req.GetTemplate(), Namespace: metav1.NamespaceAll}
	tmpl := &desktopsv1.Template{}
	if err := d.client.Get(context.TODO(), tmplnn, tmpl); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	desktop := d.newDesktopForRequest(req, sess.User.GetName())

	if err := d.client.Create(context.TODO(), desktop); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	if envTemplates := tmpl.GetEnvTemplates(); len(envTemplates) > 0 {
		var secretErr error
		defer func() {
			if secretErr != nil {
				if err := d.client.Delete(context.TODO(), desktop); err != nil {
					apiLogger.Error(err, "Couldn't cleanup desktop from failed secret creation")
				}
			}
		}()
		var data map[string][]byte
		data, secretErr = executeEnvTemplates(sess, envTemplates)
		if secretErr != nil {
			apiutil.ReturnAPIError(secretErr, w)
			return
		}
		secret := d.newEnvSecretForRequest(req, desktop, sess.User.GetName(), data)
		if secretErr = d.client.Create(context.TODO(), secret); secretErr != nil {
			apiutil.ReturnAPIError(secretErr, w)
			return
		}
	}

	apiutil.WriteJSON(&types.CreateSessionResponse{
		Name:      desktop.GetName(),
		Namespace: desktop.GetNamespace(),
	}, w)
}

func (d *desktopAPI) newDesktopForRequest(req *types.CreateSessionRequest, username string) *desktopsv1.Session {
	return &desktopsv1.Session{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", req.GetTemplate()),
			Namespace:    req.GetNamespace(),
			Labels:       d.vdiCluster.GetUserDesktopSelector(username),
		},
		Spec: desktopsv1.SessionSpec{
			VDICluster:     d.vdiCluster.GetName(),
			Template:       req.GetTemplate(),
			User:           username,
			ServiceAccount: req.GetServiceAccount(),
		},
	}
}

func (d *desktopAPI) newEnvSecretForRequest(req *types.CreateSessionRequest, desktop *desktopsv1.Session, username string, data map[string][]byte) *corev1.Secret {
	labels := desktop.GetLabels()
	labels[v1.DesktopNameLabel] = desktop.GetName()
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-env-", username),
			Namespace:    req.GetNamespace(),
			Labels:       labels,
		},
		Data: data,
	}
}

func executeEnvTemplates(sess *types.JWTClaims, envTemplates map[string]string) (map[string][]byte, error) {
	data := make(map[string][]byte)
	for envVar, envVarTmpl := range envTemplates {
		t, err := template.New("").Parse(envVarTmpl)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, map[string]interface{}{
			"Session": sess,
		}); err != nil {
			return nil, err
		}
		data[envVar] = buf.Bytes()
	}
	return data, nil
}
