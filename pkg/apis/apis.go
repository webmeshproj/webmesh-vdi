// Package apis contains the Kubernetes and internal APIs used by kVDI.
package apis

import (
	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	if err := promv1.AddToScheme(s); err != nil {
		return err
	}
	return AddToSchemes.AddToScheme(s)
}
