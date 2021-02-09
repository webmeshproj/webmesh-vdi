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

func init() {
	SchemeBuilder.Register(&VDIRole{}, &VDIRoleList{})
}
