package v1

import "time"

const (
	// RoleClusterRefLabel marks for which cluster a role belongs
	RoleClusterRefLabel = "kvdi.io/cluster-ref"
	// CreationSpecAnnotation contains the serialized creation spec of a resource
	// to be compared against desired state.
	CreationSpecAnnotation = "kvdi.io/creation-spec"
	// LDAPGroupRoleAnnotation is an annotation applied to VDIRoles to "bind" them
	// to LDAP groups. A semicolon-separated list can bind a VDIRole to multiple
	// LDAP groups.
	LDAPGroupRoleAnnotation = "kvdi.io/ldap-groups"
	// OIDCGroupRoleAnnotation is the annotation applied to VDIRoles to "bind" them
	// to groups provided in claims from an OIDC provider. A semicolon separated list can
	// bind a role to multiple groups.
	OIDCGroupRoleAnnotation = "kvdi.io/oidc-groups"
	// AuthGroupSeparator is the separator used when parsing lists of groups from a string.
	AuthGroupSeparator = ";"
	// VDIClusterLabel is the label attached to resources to reference their parents VDI cluster
	VDIClusterLabel = "vdiCluster"
	// ComponentLabel is the label primarily used for service selectors
	ComponentLabel = "vdiComponent"
	// UserLabel is a label to tie the user id associated with a desktop instance
	UserLabel = "desktopUser"
	// DesktopNameLabel is a label referencing the name of the desktop instance. This is to add randomness
	// for the headless service selector placed in front of each pod.
	DesktopNameLabel = "desktopName"
	// ClientAddrLabel is the a label referencing the client address on a display/audio lock.
	ClientAddrLabel = "clientAddr"
	// ServerCertificateMountPath is where server certificates get placed inside pods
	ServerCertificateMountPath = "/etc/kvdi/tls/server"
	// ClientCertificateMountPath is where client certificates get placed inside pods
	ClientCertificateMountPath = "/etc/kvdi/tls/client"
	// SecretAssetsMountPath is a mount path for assets backed by secrets
	SecretAssetsMountPath = "/etc/kvdi/secrets"
	// JWTSecretKey is where our JWT secret is stored in the secrets backend.
	JWTSecretKey = "jwtSecret"
	// OTPUsersSecretKey is where a mapping of users to their OTP secrets is held in the secrets backend.
	OTPUsersSecretKey = "otpUsers"
	// RefreshTokensSecretKey is where a mapping of refresh tokens to users is kept in the secrets backend.
	RefreshTokensSecretKey = "refreshTokens"
	// WebPort is the port that web services will listen on internally
	WebPort = 8443
	// PublicWebPort is the port for the app service
	PublicWebPort = 443
	// DesktopRunDir is the dir mounted for internal runtime files
	DesktopRunDir = "/var/run/kvdi"
	// DefaultDisplaySocketAddr is the default path used for the display unix socket
	DefaultDisplaySocketAddr = "unix:///var/run/kvdi/display.sock"
	// DefaultNamespace is the default namespace to provision resources in
	DefaultNamespace = "default"
	// DefaultSessionLength is the session length used for setting expiry
	// times on new user sessions.
	DefaultSessionLength = time.Duration(15) * time.Minute
	// CACertKey is the key where the CA certificate is placed in TLS secrets.
	CACertKey = "ca.crt"
)

// NamespaceAll represents all namespaces
const NamespaceAll = "*"

// Resource represents the target of an API action
type Resource string

// Resource options
const (
	// ResourceUsers represents users of kVDI. This action would only apply
	// when using local auth.
	ResourceUsers Resource = "users"
	// ResourceRoles represents the auth roles in kVDI. This would allow a user
	// to manipulate policies via the app API.
	ResourceRoles Resource = "roles"
	// ResourceTeemplates represents desktop templates in kVDI. Mainly the ability
	// to launch seessions from them and connect to them.
	ResourceTemplates Resource = "templates"
	// ResourceAll matches all resources
	ResourceAll Resource = "*"
)

// Verb represents an API action
type Verb string

// Verb options
const (
	// Create operations
	VerbCreate Verb = "create"
	// Read operations
	VerbRead Verb = "read"
	// Update operations
	VerbUpdate Verb = "update"
	// Delete operations
	VerbDelete Verb = "delete"
	// Use operations
	VerbUse Verb = "use"
	// Launch operations
	VerbLaunch Verb = "launch"
	// VerbAll matches all actions
	VerbAll Verb = "*"
)

// Other defaults that we need to the address of
var (
	DefaultUser     int64 = 1000
	DefaultReplicas int32 = 1
	TrueVal               = true
	FalseVal              = false
)
