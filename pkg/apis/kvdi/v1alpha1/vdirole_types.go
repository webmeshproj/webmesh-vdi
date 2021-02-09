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

package v1alpha1

import (
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VDIRole is the Schema for the vdiroles API
// +kubebuilder:resource:path=vdiroles,scope=Cluster
type VDIRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// A list of rules granting access to resources in the VDICluster.
	Rules []v1.Rule `json:"rules,omitempty"`
}

// GetRules returns the rules for this VDIRole.
func (v *VDIRole) GetRules() []v1.Rule { return v.Rules }

// ToUserRole converts this VDIRole to the VDIUserRole format. The VDIUserRole is
// a condensed representation meant to be stored in JWTs.
func (v *VDIRole) ToUserRole() *v1.VDIUserRole {
	return &v1.VDIUserRole{
		Name:  v.GetName(),
		Rules: v.GetRules(),
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VDIRoleList contains a list of VDIRole
type VDIRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VDIRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VDIRole{}, &VDIRoleList{})
}
