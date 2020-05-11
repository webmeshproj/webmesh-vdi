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
	// The configuration for user volumes. *NOTE:* Even though the controller
	// will try to force the reclaim policy on created volumes to `Retain`, you
	// may want to set it explicitly on your storage-class controller as an extra
	// safeguard.
	UserDataSpec *corev1.PersistentVolumeClaimSpec `json:"userdataSpec,omitempty"`
	// App configurations.
	App *AppConfig `json:"app,omitempty"`
	// Authentication configurations
	Auth *AuthConfig `json:"auth,omitempty"`
	// Secrets backend configurations
	Secrets *SecretsConfig `json:"secrets,omitempty"`
}

// AppConfig represents app configurations for the VDI cluster
type AppConfig struct {
	// The image to use for the app instances. Defaults to the public image
	// matching the version of the currently running manager.
	Image string `json:"image,omitempty"`
	// Whether to add CORS headers to API requests
	CORSEnabled bool `json:"corsEnabled,omitempty"`
	// Whether to log auditing events to stdout
	AuditLog bool `json:"auditLog,omitempty"`
	// The number of app replicas to run
	Replicas int32 `json:"replicas,omitempty"`
	// Resource requirements to place on the app pods
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// AuthConfig will be for authentication driver configurations. The goal
// is to support multiple backends, e.g. local, oauth, ldap, etc.
type AuthConfig struct {
	// Allow anonymous users to create desktop instances
	AllowAnonymous bool `json:"allowAnonymous,omitempty"`
	// A secret where a generated admin password will be stored
	AdminSecret string `json:"adminSecret,omitempty"`
	// Use local auth (secret-backed) authentication
	LocalAuth *LocalAuthConfig `json:"localAuth,omitempty"`
}

// SecretsConfig configurese the backend for secrets management.
type SecretsConfig struct {
	// Use a kubernetes secret for storing sensitive values. If no other coniguration is provided
	// then this is the fallback.
	K8SSecret *K8SSecretConfig `json:"k8sSecret,omitempty"`
	// Use vault for storing sensitive values. Requires kubernetes service account
	// authentication.
	Vault *VaultConfig `json:"vault,omitempty"`
}

// LocalAuthConfig represents a local, 'passwd'-like authentication driver.
type LocalAuthConfig struct{}

// K8SSecretConfig uses a Kubernetes secret to store and retrieve sensitive values.
type K8SSecretConfig struct {
	// The name of the secret backing the values. Default is `<cluster-name>-app-secrets`.
	SecretName string `json:"secretName,omitempty"`
}

// VaultConfig represents the configurations for connecting to a vault server.
type VaultConfig struct {
	// The full URL to the vault server. Same as the `VAULT_ADDR` variable.
	Address string `json:"address"`
	// The base64 encoded CA certificate for verifying the vault server certificate.
	CACertBase64 string `json:"caCertBase64,omitempty"`
	// Set to true to disable TLS verification.
	Insecure bool `json:"insecure,omitempty"`
	// Optionally set the SNI when connecting using HTTPS.
	TLSServerName string `json:"tlsServerName,omitempty"`
	// The auth role to assume when authenticating against vault. Defaults to `kvdi`.
	AuthRole string `json:"authRole,omitempty"`
	// The base path to store secrets in vault.
	SecretsPath string `json:"secretsPath,omitempty"`
}

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
