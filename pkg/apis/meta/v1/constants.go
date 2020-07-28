package v1

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
	// ServerCertificateMountPath is where server certificates get placed inside pods
	ServerCertificateMountPath = "/etc/kvdi/tls/server"
	// ClientCertificateMountPath is where client certificates get placed inside pods
	ClientCertificateMountPath = "/etc/kvdi/tls/client"
	// SecretAssetsMountPath is a mount path for assets backed by secrets
	SecretAssetsMountPath = "/etc/kvdi/secrets"
	// JWTSecretKey is where our JWT secret is stored in a secrets backend.
	JWTSecretKey = "jwtSecret"
	// OTPUsersSecretKey is where a mapping of users to their OTP secrets is held in a secrets backend.
	OTPUsersSecretKey = "otpUsers"
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
	ResourceRoles = "roles"
	// ResourceTeemplates represents desktop templates in kVDI. Mainly the ability
	// to launch seessions from them and connect to them.
	ResourceTemplates = "templates"
	// ResourceAll matches all resources
	ResourceAll = "*"
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

// Other defaults that we need to the address of
var (
	DefaultUser     int64 = 1000
	DefaultReplicas int32 = 1
	TrueVal               = true
	FalseVal              = false
)
