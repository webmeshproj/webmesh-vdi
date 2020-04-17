package auth

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/local"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

func GetAuthProvider(cluster *v1alpha1.VDICluster) apiutil.AuthProvider {
	return local.New()
}
