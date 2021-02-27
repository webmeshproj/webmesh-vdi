/*

	Copyright 2020,2021 Avi Zimmerman

	This file is part of kvdi.

	kvdi is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	kvdi is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with kvdi.  If not, see <https://www.gnu.org/licenses/>.

*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=vdiroles,scope=Cluster
//+kubebuilder:subresource:status

// VDIRole is the Schema for the vdiroles API
type VDIRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// A list of rules granting access to resources in the VDICluster.
	Rules []Rule `json:"rules,omitempty"`
}

// GetRules returns the rules for this VDIRole.
func (v *VDIRole) GetRules() []Rule { return v.Rules }

//+kubebuilder:object:root=true

// VDIRoleList contains a list of VDIRole
type VDIRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VDIRole `json:"items"`
}

// Trim will trim the managed fields and other metadata not used in processing. It
// has the benefit of producing less data when sending over the wire. Note that the
// objects returned by this method should NOT be used when sending later Update requests.
func (v *VDIRoleList) Trim() []*VDIRole {
	if len(v.Items) == 0 {
		return nil
	}
	out := make([]*VDIRole, len(v.Items))
	for i, role := range v.Items {
		r := role.DeepCopy()
		r.SetManagedFields(nil)
		r.SetOwnerReferences(nil)
		r.SetGeneration(0)
		r.SetResourceVersion("")
		r.SetUID(types.UID(""))
		if annotations := r.GetAnnotations(); annotations != nil {
			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
			r.SetAnnotations(annotations)
		}
		out[i] = r
	}
	return out
}

func init() {
	SchemeBuilder.Register(&VDIRole{}, &VDIRoleList{})
}
