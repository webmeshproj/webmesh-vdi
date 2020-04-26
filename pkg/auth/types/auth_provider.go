package types

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
)

// AuthProvider defines an interface for handling login attempts. Currently
// only Local auth (db-based) is supported, however other integrations such as
// LDAP or OAuth can implement this interface.
type AuthProvider interface {
	Setup(*v1alpha1.VDICluster) error
	Authenticate(w http.ResponseWriter, r *http.Request)
}
