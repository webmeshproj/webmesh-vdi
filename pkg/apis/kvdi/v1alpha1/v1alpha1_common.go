package v1alpha1

// Annotations used for internal operations on resources
const (
	// CreationSpecAnnotation contains the serialized creation spec of a resource
	// to be compared against desired state.
	CreationSpecAnnotation = "kvdi.io/creation-spec"
	// The label attached to resources to reference their parents VDI cluster
	VDIClusterLabel = "vdiCluster"
	// The component label primarily used for service selectors
	ComponentLabel = "vdiComponent"
	// A label to tie the user id associated with a desktop instance
	UserLabel = "desktopUser"
	// A label referencing the name of the desktop instance. This is to add randomness
	// for the headless service selector placed in front of each pod.
	DesktopNameLabel = "desktopName"
	// Where server certificates get placed inside pods
	ServerCertificateMountPath = "/etc/kvdi/tls/server"
	// Where client certificates get placed inside pods
	ClientCertificateMountPath = "/etc/kvdi/tls/client"
	// A mount path for assets backed by secrets
	SecretAssetsMountPath = "/etc/kvdi/secrets"
	// Where our JWT secret is stored in a secrets backend.
	JWTSecretKey = "jwtSecret"
	// Where a mapping of users to their OTP secrets is held in a secrets backend.
	OTPUsersSecretKey = "otpUsers"
	// The port that web servicees will listen on internally
	WebPort = 8443
	// The port for the app service
	PublicWebPort = 443
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
