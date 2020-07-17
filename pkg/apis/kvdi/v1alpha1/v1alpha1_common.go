package v1alpha1

const (
	// CreationSpecAnnotation contains the serialized creation spec of a resource
	// to be compared against desired state.
	CreationSpecAnnotation = "kvdi.io/creation-spec"
	// LDAPGroupRoleAnnotation is an annotation applied to VDIRoles to "bind" them
	// to LDAP groups. A semicolon-separated list can bind a VDIRole to multiple
	// LDAP groups.
	LDAPGroupRoleAnnotation = "kvdi.io/ldap-groups"
	// LDAPGroupSeparator is the separator used when parsing lists of groups from a string.
	LDAPGroupSeparator = ";"
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
)

// Defaults
const (
	// defaultNamespace is the default namespace to provision resources in
	defaultNamespace string = "default"
)

// Other defaults that we need to the address of
var (
	defaultUser     int64 = 1000
	defaultReplicas int32 = 1
	trueVal               = true
	falseVal              = false
)
