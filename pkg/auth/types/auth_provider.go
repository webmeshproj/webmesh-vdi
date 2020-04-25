package types

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
)

type AuthProvider interface {
	Setup(*v1alpha1.VDICluster) error
	Authenticate(w http.ResponseWriter, r *http.Request)
}
