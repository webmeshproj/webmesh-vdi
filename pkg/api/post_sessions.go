package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Request for a new desktop session
// swagger:parameters postSessionRequest
type swaggerCreateSessionRequest struct {
	// in:body
	Body v1.CreateSessionRequest
}

// CreateSessionResponse returns the name of the Desktop and what namespace
// it is running in.
type CreateSessionResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// New session response
// swagger:response postSessionResponse
type swaggerCreateSessionResponse struct {
	// in:body
	Body CreateSessionResponse
}

// swagger:route POST /api/sessions Sessions postSessionRequest
// Creates a new desktop session with the given parameters.
// responses:
//   200: postSessionResponse
//   400: error
//   403: error
func (d *desktopAPI) StartDesktopSession(w http.ResponseWriter, r *http.Request) {
	sess := apiutil.GetRequestUserSession(r)
	req := apiutil.GetRequestObject(r).(*v1.CreateSessionRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	if max := d.vdiCluster.GetMaxSessionsPerUser(); max > 0 {
		desktops := &v1alpha1.DesktopList{}
		if err := d.client.List(context.TODO(), desktops, client.InNamespace(metav1.NamespaceAll), client.MatchingLabels(d.vdiCluster.GetUserDesktopSelector(sess.User.Name))); err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		if len(desktops.Items) >= max {
			apiutil.ReturnAPIError(fmt.Errorf("%s has reached the maximum allowed (%d) running desktops", sess.User.Name, max), w)
			return
		}
	}

	tmplnn := types.NamespacedName{Name: req.GetTemplate(), Namespace: metav1.NamespaceAll}
	tmpl := &v1alpha1.DesktopTemplate{}
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

	apiutil.WriteJSON(&CreateSessionResponse{
		Name:      desktop.GetName(),
		Namespace: desktop.GetNamespace(),
	}, w)
}

func (d *desktopAPI) newDesktopForRequest(req *v1.CreateSessionRequest, username string) *v1alpha1.Desktop {
	return &v1alpha1.Desktop{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", req.GetTemplate()),
			Namespace:    req.GetNamespace(),
			Labels:       d.vdiCluster.GetUserDesktopSelector(username),
		},
		Spec: v1alpha1.DesktopSpec{
			VDICluster:     d.vdiCluster.GetName(),
			Template:       req.GetTemplate(),
			User:           username,
			ServiceAccount: req.GetServiceAccount(),
		},
	}
}

func (d *desktopAPI) newEnvSecretForRequest(req *v1.CreateSessionRequest, desktop *v1alpha1.Desktop, username string, data map[string][]byte) *corev1.Secret {
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

func executeEnvTemplates(sess *v1.JWTClaims, envTemplates map[string]string) (map[string][]byte, error) {
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
