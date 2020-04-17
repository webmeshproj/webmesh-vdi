package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PostSessionsRequest struct {
	Template  string `json:"template"`
	Namespace string `json:"namespace,omitempty"`
	User      string `json:"user,omitempty"`
}

func (p *PostSessionsRequest) GetTemplate() string {
	return p.Template
}

func (p *PostSessionsRequest) GetNamespace() string {
	if p.Namespace != "" {
		return p.Namespace
	}
	return "default"
}

func (p *PostSessionsRequest) GetUser() string {
	if p.User != "" {
		return p.User
	}
	return "anonymous"
}

type PostSessionsResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Endpoint  string `json:"endpoint"`
}

func (d *desktopAPI) StartDesktopSession(w http.ResponseWriter, r *http.Request) {
	req := PostSessionsRequest{}
	if err := apiutil.UnmarshalRequest(r, &req); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	if req.Template == "" {
		apiutil.ReturnAPIError(errors.New("No DesktopTemplate included in the request"), w)
		return
	}

	desktop := d.newDesktopForRequest(req)
	if err := d.client.Create(context.TODO(), desktop); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(&PostSessionsResponse{
		Name:      desktop.Name,
		Namespace: desktop.Namespace,
		Endpoint:  util.DesktopShortURL(desktop),
	}, w)
}

func (d *desktopAPI) newDesktopForRequest(req PostSessionsRequest) *v1alpha1.Desktop {
	return &v1alpha1.Desktop{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", req.GetTemplate(), strings.Split(uuid.New().String(), "-")[0]),
			Namespace: req.GetNamespace(),
		},
		Spec: v1alpha1.DesktopSpec{
			VDICluster: d.vdiCluster.GetName(),
			Template:   req.GetTemplate(),
			User:       req.GetUser(),
		},
	}
}
