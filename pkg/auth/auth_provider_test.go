package auth

import (
	"reflect"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/providers/local"
)

func TestGetAuthProvider(t *testing.T) {
	authProvider := GetAuthProvider(&v1alpha1.VDICluster{}, nil)
	if reflect.TypeOf(authProvider) != reflect.TypeOf(&local.AuthProvider{}) {
		t.Error("Should have received a local auth provider")
	}
}
