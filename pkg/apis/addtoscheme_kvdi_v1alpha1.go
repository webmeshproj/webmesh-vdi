package apis

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, cm.SchemeBuilder.AddToScheme)
}
