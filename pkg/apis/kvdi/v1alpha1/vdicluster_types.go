package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VDIClusterSpec defines the desired state of VDICluster
type VDIClusterSpec struct {
	// The namespace to provision application resurces in. Defaults to the `default`
	// namespace
	AppNamespace string `json:"appNamespace,omitempty"`
	// Pull secrets to use when pulling container images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// The namespace cert-manager is running in. Defaults to `cert-manager`.
	CertManagerNamespace string `json:"certManagerNamespace,omitempty"`
	// App configurations.
	App *AppConfig `json:"app,omitempty"`
	// Authentication configurations
	Auth *AuthConfig `json:"auth,omitempty"`
	// RethinkDB configurations
	RethinkDB *RethinkDBConfig `json:"rethinkdb,omitempty"`
}

// AppConfig represents app configurations for the VDI cluster
type AppConfig struct {
	// An exterenal host name that will be used for any routes that need to be
	// broadcasted to the end user.
	ExternalHostname string `json:"externalHostname,omitempty"`
	// Whether to add CORS headers to API requests
	CORSEnabled bool `json:"corsEnabled,omitempty"`
	// The number of app replicas to run
	Replicas int32 `json:"replicas,omitempty"`
	// Resource requirements to place on the app pods
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// RethinkDBConfig represents rethinkdb configurations for the VDI cluster
type RethinkDBConfig struct {
	// The image to use for the rethinkdb instances. Defaults to rethinkdb:2.4.
	Image string `json:"image,omitempty"`
	// The spec for persistent volumes attached to the reethinkdb nodes
	PVCSpec *corev1.PersistentVolumeClaimSpec `json:"pvcSpec,omitempty"`
	// The number of shards to create for each table in the database.
	Shards int32 `json:"shards,omitempty"`
	// The number of data rpelicas to run for each table.
	Replicas int32 `json:"replicas,omitempty"`
	// The number of proxy instances to run.
	ProxyReplicas int32 `json:"proxyReplicas,omitempty"`
	// Resource requirements for the database pods.
	DBResources corev1.ResourceRequirements `json:"dbResources,omitempty"`
	// Resource requirements forr the proxy pods.
	ProxyResources corev1.ResourceRequirements `json:"proxyResources,omitempty"`
}

// AuthConfig will be for authentication driver configurations. The goal
// is to support multiple backends, e.g. local, oauth, ldap, etc.
type AuthConfig struct {
	// Allow anonymous users to create desktop instances
	AllowAnonymous bool `json:"allowAnonymous,omitempty"`
	// A secret where a generated admin password will be stored
	AdminSecret string `json:"adminSecret,omitempty"`
	// Use local auth (db-backed) authentication
	LocalAuth *LocalAuthConfig `json:"localAuth,omitempty"`
}

// LocalAuthConfig represents a local, db-based authentication driver.
type LocalAuthConfig struct{}

// VDIClusterStatus defines the observed state of VDICluster
type VDIClusterStatus struct {
	Ready bool `json:"ready,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VDICluster is the Schema for the vdiclusters API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=vdiclusters,scope=Cluster
type VDICluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VDIClusterSpec   `json:"spec,omitempty"`
	Status VDIClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VDIClusterList contains a list of VDICluster
type VDIClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VDICluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VDICluster{}, &VDIClusterList{})
}
