package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Labels used for configuring how the role gets applied
const (
	// RoleClusterRefLabel marks for which cluster a role belongs
	RoleClusterRefLabel = "kvdi.io/cluster-ref"
	// Other labels could be used to pass auth provider level configurations,
	// such as mapping users to roles.
)

// Verb represents an API action
type Verb string

// Verb options
const (
	// Create operations
	VerbCreate Verb = "create"
	// Read operations
	VerbRead = "read"
	// Update operations
	VerbUpdate = "update"
	// Delete operations
	VerbDelete = "delete"
	// Use operations
	VerbUse = "use"
	// Launch operations
	VerbLaunch = "launch"
	// VerbAll matches all actions
	VerbAll = "*"
)

// Resource represents the target of an API action
type Resource string

// Resource options
const (
	// ResourceUsers represents users of kVDI. This action would only apply
	// when using local auth.
	ResourceUsers Resource = "users"
	// ResourceRoles represents the auth roles in kVDI. This would allow a user
	// to manipulate policies via the app API.
	ResourceRoles = "roles"
	// ResourceTeemplates represents desktop templates in kVDI. Mainly the ability
	// to launch seessions from them and connect to them.
	ResourceTemplates = "templates"
	// ResourceAll matches all resources
	ResourceAll = "*"
)

const NamespaceAll = "*"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VDIRole is the Schema for the vdiroles API
// +kubebuilder:resource:path=vdiroles,scope=Cluster
type VDIRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// A list of rules granting access to resources in the VDICluster.
	Rules []Rule `json:"rules,omitempty"`
}

func (v *VDIRole) GetRules() []Rule { return v.Rules }

func (v *VDIRole) ToUserRole() *VDIUserRole {
	return &VDIUserRole{
		Name:  v.GetName(),
		Rules: v.GetRules(),
	}
}

// Rule represents a set of permissions applied to a VDIRole. It mostly resembles
// an rbacv1.PolicyRule, with resources being a regex and the addition of a
// namespace selector.
type Rule struct {
	// The actions this rule applies for. VerbAll matches all actions.
	Verbs []Verb `json:"verbs,omitempty"`
	// Resources this rule applies to. ResourceAll matches all resources.
	Resources []Resource `json:"resources,omitempty"`
	// Resource regexes that match this rule. This can be template patterns, role
	// names or user names. There is no All representation because * will have
	// that effect on its own when the regex is evaluated.
	ResourcePatterns []string `json:"resourcePatterns,omitempty"`
	// Namespaces this rule applies to. Only evaluated for template launching
	// permissions. NamespaceAll matches all namespaces.
	Namespaces []string `json:"namespaces,omitempty"`
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
