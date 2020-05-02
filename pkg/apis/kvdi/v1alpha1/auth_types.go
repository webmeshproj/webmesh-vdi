package v1alpha1

import (
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthProvider defines an interface for handling login attempts. Currently
// only Local auth (db-based) is supported, however other integrations such as
// LDAP or OAuth can implement this interface.
type AuthProvider interface {
	// Reconcile should ensure any k8s resources required for this authentication
	// provider.
	Reconcile(logr.Logger, client.Client, *VDICluster, string) error
	// Setup is called when the kVDI app launches and is a chance for the provider
	// to setup any resources it needs to serve requests.
	Setup(client.Client, *VDICluster) error

	// HTTP methods
	// Not all providers will be able to implement all of these methods. When
	// they can't they should serve a concise error message as to why.

	// Authenticate is called for API authentication requests. It should generate
	// a new JWTClaims object and serve a SessionResponse back to the user.
	Authenticate(*LoginRequest) (*AuthResult, error)
	// GetUsers should return a list of VDIUsers.
	GetUsers() ([]*VDIUser, error)
	// GetUser should retrieve a single VDIUser.
	GetUser(string) (*VDIUser, error)
	// PostUser should handle any logic required to register a new user in kVDI.
	CreateUser(*CreateUserRequest) error
	// PutUser should update a VDIUser.
	UpdateUser(string, *UpdateUserRequest) error
	// DeleteUser should remove a VDIUser.
	DeleteUser(string) error
}

// AuthResult represents a response from an authentication attempt to a provider.
// It contains user information, roles, and any other auth requirements.
type AuthResult struct {
	// The authenticated user and their roles
	User *VDIUser
}

// JWTClaims represents the claims used when issuing JWT tokens.
type JWTClaims struct {
	// The user with their permissions when the token was generated
	User *VDIUser `json:"user"`
	// Whether the user is fully authorized
	Authorized bool `json:"authorized"`
	// The standard JWT claims
	jwt.StandardClaims
}

const (
	// DefaultSessionLength is the session length used for setting expiry
	// times on new user sessions.
	DefaultSessionLength = time.Duration(8) * time.Hour
)

// VDIUser represents a user in kVDI. It is the auth providers responsibility
// to take an authentication request and generate a JWT with claims defining
// this object.
type VDIUser struct {
	VDIRole `json:"-"`
	// A unique name for the user
	Name string `json:"name"`
	// A list of roles applide to the user. The grants associated with each user
	// are embedded in the JWT signed when authenticating.
	Roles []*VDIUserRole `json:"roles"`
	// Whether or not MFA is enabled for this user
	MFAEnabled bool `json:"mfaEnabled"`
}

// VDIUserRole represents a VDIRole, but only with the data that is to be
// embedded in the JWT. Primarily, leaving out useless metadata that will inflate
// the token.
type VDIUserRole struct {
	// The name of the role, this must match the VDIRole from which this object
	// derives.
	Name string `json:"name"`
	// The rules for this role.
	Rules []Rule `json:"rules"`
}

// APIAction represents an API action to evaluate against a user's roles.
type APIAction struct {
	// The verb type of the action
	Verb Verb `json:"verb"`
	// The resource type of the action
	ResourceType Resource `json:"resourceType"`
	// The name of the targeted resource
	ResourceName string `json:"resourceName"`
	// The namespace of the targeted resource
	ResourceNamespace string `json:"resourceNamespace,omitempty"`
}

// ResourceNameString returns a user friendly resource name string
func (a *APIAction) ResourceNameString() string {
	if a.ResourceNamespace != "" && a.ResourceName != "" {
		return fmt.Sprintf("%s/%s", a.ResourceNamespace, a.ResourceName)
	}
	if a.ResourceName != "" {
		return a.ResourceName
	}
	if a.ResourceNamespace != "" {
		return a.ResourceNamespace
	}
	return ""
}

// String returns a user friendly string describing the action
func (a *APIAction) String() string {
	if a.Verb == "" && a.ResourceType == "" {
		return ""
	}
	str := fmt.Sprintf("%s %s", strings.ToUpper(string(a.Verb)), strings.Title(string(a.ResourceType)))
	if resourceName := a.ResourceNameString(); resourceName != "" {
		str = str + fmt.Sprintf(" %s", resourceName)
	}
	return str
}

// ResourceGetter is an interface for retrieving lists of kVDI related resources.
// Its primary purpose is to pass an interface to rbac evaluations so they can
// check permissions against present resources.
type ResourceGetter interface {
	// Retrieves DesktopTemplates
	TemplatesGetter
	// Retrieves VDIUsers
	UsersGetter
	// Retrieves VDIRoles
	RolesGetter
}

// TemplatesGetter is an interface that can be used to retrieve available
// templates while checking user permissions.
type TemplatesGetter interface {
	GetTemplates() ([]DesktopTemplate, error)
}

// UsersGetter is an interface that can be used to retrieve available
// users while checking user permissions.
type UsersGetter interface {
	GetUsers() ([]VDIUser, error)
}

// RolesGetter is an interface that can be used to retrieve available
// roles while checking user permissions.
type RolesGetter interface {
	GetRoles() ([]VDIRole, error)
}
