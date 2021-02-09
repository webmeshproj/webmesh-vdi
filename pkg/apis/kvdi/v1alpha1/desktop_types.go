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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DesktopSpec defines the desired state of Desktop
type DesktopSpec struct {
	// The VDICluster this Desktop belongs to. This helps to determine which app
	// instance certificates need to be created for.
	VDICluster string `json:"vdiCluster"`
	// The DesktopTemplate for booting this instance.
	Template string `json:"template"`
	// The username to use inside the instance, defaults to `anonymous`.
	User string `json:"user,omitempty"`
	// A service account to tie to the pod for this instance.
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// DesktopStatus defines the observed state of Desktop
type DesktopStatus struct {
	// Whether the instance is running and resolvable within the cluster.
	Running  bool            `json:"running,omitempty"`
	PodPhase corev1.PodPhase `json:"podPhase,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Desktop is the Schema for the desktops API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=desktops,scope=Namespaced
// +kubebuilder:printcolumn:name="User",type="string",JSONPath=".spec.user"
// +kubebuilder:printcolumn:name="ServiceAccount",type="string",JSONPath=".spec.serviceAccount"
// +kubebuilder:printcolumn:name="Template",type="string",JSONPath=".spec.template"
type Desktop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DesktopSpec   `json:"spec,omitempty"`
	Status DesktopStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DesktopList contains a list of Desktop
type DesktopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Desktop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Desktop{}, &DesktopList{})
}
