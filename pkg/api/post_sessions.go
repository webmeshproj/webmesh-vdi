package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Request for a new desktop session
// swagger:parameters postSessionRequest
type swaggerCreateSessionRequest struct {
	// in:body
	Body v1alpha1.CreateSessionRequest
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

// swagger:route POST /api/sessions Desktops postSessionRequest
// Creates a new desktop session with the given parameters.
// responses:
//   200: postSessionResponse
//   400: error
//   403: error
func (d *desktopAPI) StartDesktopSession(w http.ResponseWriter, r *http.Request) {
	sess := apiutil.GetRequestUserSession(r)
	req := apiutil.GetRequestObject(r).(*v1alpha1.CreateSessionRequest)
	if req == nil {
		apiutil.ReturnAPIError(errors.New("Malformed request"), w)
		return
	}

	desktop := d.newDesktopForRequest(req, sess.User.GetName())

	if err := d.client.Create(context.TODO(), desktop); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&CreateSessionResponse{
		Name:      desktop.GetName(),
		Namespace: desktop.GetNamespace(),
	}, w)
}

// newDesktopForRequest builds a new Desktop object from the request parameters.
func (d *desktopAPI) newDesktopForRequest(req *v1alpha1.CreateSessionRequest, username string) *v1alpha1.Desktop {
	return &v1alpha1.Desktop{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", req.GetTemplate(), strings.Split(uuid.New().String(), "-")[0]),
			Namespace: req.GetNamespace(),
			Labels:    d.vdiCluster.GetUserDesktopLabels(username),
		},
		Spec: v1alpha1.DesktopSpec{
			VDICluster: d.vdiCluster.GetName(),
			Template:   req.GetTemplate(),
			User:       username,
		},
	}
}
