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
	// The configuration for user volumes. **NOTE:** Even though the controller
	// will try to force the reclaim policy on created volumes to `Retain`, you
	// may want to set it explicitly on your storage-class controller as an extra
	// safeguard.
	UserDataSpec *corev1.PersistentVolumeClaimSpec `json:"userdataSpec,omitempty"`
	// App configurations.
	App *AppConfig `json:"app,omitempty"`
	// Authentication configurations
	Auth *AuthConfig `json:"auth,omitempty"`
	// Global desktop configurations
	Desktops *DesktopsConfig `json:"desktops,omitempty"`
	// Secrets backend configurations
	Secrets *SecretsConfig `json:"secrets,omitempty"`
	// Metrics configurations.
	Metrics *MetricsConfig `json:"metrics,omitempty"`
}

// DesktopsConfig represents global configurations for desktop
// sessions.
type DesktopsConfig struct {
	// When configured, desktop sessions will be forcefully terminated when
	// the time limit is reached.
	MaxSessionLength string `json:"maxSessionLength,omitempty"`
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
	// The type of service to create in front of the app instance.
	// Defaults to `LoadBalancer`.
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Extra annotations to apply to the app service.
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	// TLS configurations for the app instance
	TLS *TLSConfig `json:"tls,omitempty"`
	// Resource requirements to place on the app pods
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// TLSConfig contains TLS configurations for kVDI.
type TLSConfig struct {
	// A pre-existing TLS secret to use for the HTTPS listener. If not defined,
	// a certificate is generated.
	ServerSecret string `json:"serverSecret,omitempty"`
}

// MetricsConfig contains configuration options for gathering metrics.
type MetricsConfig struct {
	// Configurations for creating a ServiceMonitor CR for a pre-existing
	// prometheus-operator installation.
	ServiceMonitor *ServiceMonitorConfig `json:"serviceMonitor,omitempty"`
	// Prometheus deployment configurations.
	// **NOT IMPLEMENTED:** Toying with the idea of having the manager deploy
	// a prometheus instance for scraping.
	Prometheus *PrometheusConfig `json:"prometheus,omitempty"`
	// Grafana sidecar configurations.
	// **NOT IMPLEMENTED:** In the same spirit as the prometheus configurations,
	// toying with the idea of running grafana sidecars for visualizing metrics in
	// the UI.
	Grafana *GrafanaConfig `json:"grafana,omitempty"`
}

// ServiceMonitorConfig contains configuration options for creating a ServiceMonitor.
type ServiceMonitorConfig struct {
	// Set to true to create a ServiceMonitor object for the kvdi metrics.
	Create bool `json:"create,omitempty"`
	// Extra labels to apply to the ServiceMonitor object. Set these to the selector
	// in your prometheus-operator configuration (usually `{"release": "<helm_release_name>"}`).
	// Defaults to `{"release": "prometheus"}`.
	Labels map[string]string `json:"labels,omitempty"`
}

// PrometheusConfig contains configuration options for a prometheus deployment.
type PrometheusConfig struct {
	// Set to true to create a prometheus instance.
	Create bool `json:"create,omitempty"`
	// Resource requirements to place on the Prometheus deployment
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// GrafanaConfig contains configuration options for the grafana sidecar.
type GrafanaConfig struct {
	// Set to true to run a grafana sidecar with the app pods. This can be used to visualize
	// data in the prometheus deployment.
	Enabled bool `json:"enabled,omitempty"`
}

// AuthConfig will be for authentication driver configurations. The goal
// is to support multiple backends, e.g. local, oauth, ldap, etc.
type AuthConfig struct {
	// Allow anonymous users to create desktop instances
	AllowAnonymous bool `json:"allowAnonymous,omitempty"`
	// A secret where a generated admin password will be stored
	AdminSecret string `json:"adminSecret,omitempty"`
	// How long issued access tokens should be valid for. When using OIDC auth
	// you may want to set this to a higher value (e.g. 8-10h) since the refresh token
	// flow will not be able to lookup a user's grants from the provider. Defaults to `15m`.
	TokenDuration string `json:"tokenDuration,omitempty"`
	// Use local auth (secret-backed) authentication
	LocalAuth *LocalAuthConfig `json:"localAuth,omitempty"`
	// Use LDAP for authentication.
	LDAPAuth *LDAPConfig `json:"ldapAuth,omitempty"`
	// Use OIDC for authentication
	OIDCAuth *OIDCConfig `json:"oidcAuth,omitempty"`
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

// LDAPConfig represents the configurations for using LDAP as the authentication
// backend.
type LDAPConfig struct {
	// The URL to the LDAP server.
	URL string `json:"url,omitempty"`
	// Set to true to skip TLS verification of an `ldaps` connection.
	TLSInsecureSkipVerify bool `json:"tlsInsecureSkipVerify,omitempty"`
	// The base64 encoded CA certificate to use when verifying the TLS certificate of
	// the LDAP server.
	TLSCACert string `json:"tlsCACert,omitempty"`
	// If you want to use the built-in secrets backend (vault or k8s currently),
	// set this to either the name of the secret in the vault path (the key must be "data" for now), or the key of
	// the secret used in `secrets.k8sSecret.secretName`. In default configurations this is
	// `kvdi-app-secrets`. Defaults to `ldap-userdn`.
	BindUserDNSecretKey string `json:"bindUserDNSecretKey,omitempty"`
	// Similar to the `bindUserDNSecretKey`, but for the location of the password
	// secret. Defaults to `ldap-password`.
	BindPasswordSecretKey string `json:"bindPasswordSecretKey,omitempty"`
	// If you'd rather create a separate k8s secret (instead of the configured backend)
	// for the LDAP credentials, set its name here. The keys in the secret need to
	// be defined in the other fields still. Default is to use the secret backend.
	BindCredentialsSecret string `json:"bindCredentialsSecret,omitempty"`
	// Group DNs that are allowed administrator access to the cluster. Kubernetes
	// admins will still have the ability to change configurations via the CRDs.
	AdminGroups []string `json:"adminGroups,omitempty"`
	// The base scope to search for users in. Default is to search the entire
	// directory.
	UserSearchBase string `json:"userSearchBase,omitempty"`
	// The user ID attribute to use when looking up a provided username. Defaults to `uid`.
	// This value may be different depending on the LDAP provider. For example, in an Active Directory
	// environment you may want to set this value to `sAMAccountName`.
	UserIDAttribute string `json:"userIDAttribute,omitempty"`
	// The user attribute use to lookup group membership in LDAP. Defaults to `memberOf`.
	UserGroupsAttribute string `json:"userGroupsAttribute,omitempty"`
}

// IsUndefined returns true if the given LDAPConfig object is not actually configured.
// It checks that required values are present.
func (l *LDAPConfig) IsUndefined() bool {
	return l.URL == ""
}

// OIDCConfig represents configurations for using an OIDC/OAuth provider for
// authentication.
type OIDCConfig struct {
	// The OIDC issuer URL used for discovery
	IssuerURL string `json:"issuerURL,omitempty"`
	// When using the built-in secrets backend, the key to where the client-id is
	// stored. Set this to either the name of the secret in the vault path (the key must be "data" for now),
	// or the key of the secret used in `secrets.k8sSecret.secretName`. When configuring `clientCredentialsSecret`,
	// set this to the key in that secret. Defaults to `oidc-clientid`.
	ClientIDKey string `json:"clientIDKey,omitempty"`
	// Similar to `clientIDKey`, but for the location of the client secret. Defaults
	// to `oidc-clientsecret`.
	ClientSecretKey string `json:"clientSecretKey,omitempty"`
	// When creating your own kubernets secret with the `clientIDKey` and `clientSecretKey`,
	// set this to the name of the created secret. It must be in the same namespace
	// as the manager and app instances. Defaults to `oidc-clientsecret`.
	ClientCredentialsSecret string `json:"clientCredentialsSecret,omitempty"`
	// The redirect URL path configured in the OIDC provider. This should be the full
	// path where kvdi is hosted followed by `/api/login`. For example, if `kvdi` is
	// hosted at https://kvdi.local, then this value should be set `https://kvdi.local/api/login`.
	RedirectURL string `json:"redirectURL,omitempty"`
	// The scopes to request with the authentication request. Defaults to
	// `["openid", "email", "profile", "groups"]`.
	Scopes []string `json:"scopes,omitempty"`
	// If your OIDC provider does not return a `groups` object, set this to the user
	// attribute to use for binding authenticated users to VDIRoles. Defaults to `groups`.
	GroupScope string `json:"groupScope,omitempty"`
	// Groups that are allowed administrator access to the cluster. Kubernetes
	// admins will still have the ability to change rbac configurations via the CRDs.
	AdminGroups []string `json:"adminGroups,omitempty"`
	// Set to true to skip TLS verification of an OIDC provider.
	TLSInsecureSkipVerify bool `json:"tlsInsecureSkipVerify,omitempty"`
	// The base64 encoded CA certificate to use when verifying the TLS certificate of
	// the OIDC provider.
	TLSCACert string `json:"tlsCACert,omitempty"`
	// Set to true if the OIDC provider does not support the "groups" claim (or any
	// valid alternative) and/or you would like to allow any authenticated user
	// read-only access.
	AllowNonGroupedReadOnly bool `json:"allowNonGroupedReadOnly,omitempty"`
}

// IsUndefined returns true if the given OIDCConfig object is not actually configured.
// It checks that required values are present.
func (o *OIDCConfig) IsUndefined() bool {
	return o.IssuerURL == "" || o.RedirectURL == ""
}

// K8SSecretConfig uses a Kubernetes secret to store and retrieve sensitive values.
type K8SSecretConfig struct {
	// The name of the secret backing the values. Default is `<cluster-name>-app-secrets`.
	SecretName string `json:"secretName,omitempty"`
}

// VaultConfig represents the configurations for connecting to a vault server.
type VaultConfig struct {
	// The full URL to the vault server. Same as the `VAULT_ADDR` variable.
	Address string `json:"address,omitempty"`
	// The base64 encoded CA certificate for verifying the vault server certificate.
	CACertBase64 string `json:"caCertBase64,omitempty"`
	// Set to true to disable TLS verification.
	Insecure bool `json:"insecure,omitempty"`
	// Optionally set the SNI when connecting using HTTPS.
	TLSServerName string `json:"tlsServerName,omitempty"`
	// The auth role to assume when authenticating against vault. Defaults to `kvdi`.
	AuthRole string `json:"authRole,omitempty"`
	// The base path to store secrets in vault. "Keys" for other configurations in the
	// context of the vault backend can be put at `<secretsPath>/<secretKey>.data`. This
	// will change in the future to support keys inside the secret itself, instead of assuming
	// `data`.
	SecretsPath string `json:"secretsPath,omitempty"`
}

// IsUndefined returns true if the given VaultConfig object is not actually configured.
// It checks that required values are present.
func (v *VaultConfig) IsUndefined() bool {
	return v.Address == ""
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
