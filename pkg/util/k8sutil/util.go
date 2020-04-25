package k8sutil

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LookupClusterByName fetches the VDICluster with the given name
func LookupClusterByName(c client.Client, name string) (*v1alpha1.VDICluster, error) {
	found := &v1alpha1.VDICluster{}
	return found, c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: metav1.NamespaceAll}, found)
}

// IsMarkedForDeletion returns true if the given cluster is marked for deletion.
func IsMarkedForDeletion(cr *v1alpha1.VDICluster) bool {
	return cr.GetDeletionTimestamp() != nil
}

// SetCreationSpecAnnotation sets an annotation with a checksum of the desired
// spec of the object.
func SetCreationSpecAnnotation(meta *metav1.ObjectMeta, obj runtime.Object) error {
	annotations := meta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	spec, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	h := sha256.New()
	if _, err := h.Write(spec); err != nil {
		return err
	}
	annotations[v1alpha1.CreationSpecAnnotation] = fmt.Sprintf("%x", h.Sum(nil))
	meta.SetAnnotations(annotations)
	return nil
}

// CreationSpecsEqual returns true if the two objects spec annotations are equal.
func CreationSpecsEqual(m1 metav1.ObjectMeta, m2 metav1.ObjectMeta) bool {
	m1ann := m1.GetAnnotations()
	m2ann := m2.GetAnnotations()
	spec1, ok := m1ann[v1alpha1.CreationSpecAnnotation]
	if !ok {
		return false
	}
	spec2, ok := m2ann[v1alpha1.CreationSpecAnnotation]
	if !ok {
		return false
	}
	return spec1 == spec2
}
