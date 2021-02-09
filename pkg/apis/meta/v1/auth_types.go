package v1

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

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
	GetTemplates() ([]string, error)
}

// UsersGetter is an interface that can be used to retrieve available
// users while checking user permissions.
type UsersGetter interface {
	GetUsers() ([]VDIUser, error)
}

// RolesGetter is an interface that can be used to retrieve available
// roles while checking user permissions.
type RolesGetter interface {
	GetRoles() ([]VDIUserRole, error)
}

// AuthResult represents a response from an authentication attempt to a provider.
// It contains user information, roles, and any other auth requirements.
type AuthResult struct {
	// The authenticated user and their roles
	User *VDIUser
	// The provider can populate this field to signify a redirect is required,
	// e.g. for OIDC.
	RedirectURL string
	// The provider can supply additional data to encode into the generated JWT.
	Data map[string]string
	// In the case of OIDC, the refresh tokens cannot be used. Because when the user
	// tries to use them, there is no way to query the provider for the user's information
	// without initializing a new auth flow. For now, the provider can set this to false to
	// signal to the server that a refresh is not possible.
	RefreshNotSupported bool
}

// JWTClaims represents the claims used when issuing JWT tokens.
type JWTClaims struct {
	// The user with their permissions when the token was generated
	User *VDIUser `json:"user"`
	// Whether the user is fully authorized
	Authorized bool `json:"authorized"`
	// Whether a refresh token was issued with the claims
	Renewable bool `json:"renewable"`
	// Additional data that was provided by the authentication provider
	Data map[string]string `json:"data"`
	// The standard JWT claims
	jwt.StandardClaims
}

// VDIUser represents a user in kVDI. It is the auth providers responsibility
// to take an authentication request and generate a JWT with claims defining
// this object.
type VDIUser struct {
	// A unique name for the user
	Name string `json:"name"`
	// A list of roles applide to the user. The grants associated with each user
	// are embedded in the JWT signed when authenticating.
	Roles []*VDIUserRole `json:"roles"`
	// MFA status for the user
	MFA *UserMFAStatus `json:"mfa"`
	// Any active sessions for the user - new field that is only populated on a
	// /api/whoami request.
	Sessions []*DesktopSession `json:"sessions,omitempty"`
}

// UserMFAStatus contains information about the MFA configurations
// for the user.
type UserMFAStatus struct {
	Enabled  bool `json:"enabled"`
	Verified bool `json:"verified"`
}

// GetName returns the name of a VDIUser.
func (u *VDIUser) GetName() string { return u.Name }

// Evaluate will iterate the user's roles and return true if any of them have
// a rule that allows the given action.
func (u *VDIUser) Evaluate(action *APIAction) bool {
	for _, role := range u.Roles {
		if ok := role.Evaluate(action); ok {
			return true
		}
	}
	return false
}

// IncludesRule returns true if the rules applied to this user are not elevated
// by any of the permissions in the provided rule.
func (u *VDIUser) IncludesRule(ruleToCheck Rule, resourceGetter ResourceGetter) bool {
	for _, role := range u.Roles {
		if ok := role.IncludesRule(ruleToCheck, resourceGetter); ok {
			return true
		}
	}
	return false
}

// FilterNamespaces will take a list of namespaces and filter them based off
// the ones this user can provision desktops in.
func (u *VDIUser) FilterNamespaces(nss []string) []string {
	filtered := make([]string, 0)
	for _, ns := range nss {
		action := &APIAction{
			Verb:              VerbLaunch,
			ResourceType:      ResourceTemplates,
			ResourceNamespace: ns,
		}
		if u.Evaluate(action) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

// FilterServiceAccounts will take a list of service accounts and a given namespace,
// and filter them based off the ones this user can assume with desktops.
func (u *VDIUser) FilterServiceAccounts(sas []string, ns string) []string {
	filtered := make([]string, 0)
	for _, sa := range sas {
		action := &APIAction{
			Verb:              VerbUse,
			ResourceType:      ResourceServiceAccounts,
			ResourceName:      sa,
			ResourceNamespace: ns,
		}
		if u.Evaluate(action) {
			filtered = append(filtered, sa)
		}
	}
	return filtered
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

// GetName returns the name of the role
func (r *VDIUserRole) GetName() string { return r.Name }

// Evaluate iterates all the rules in this role and returns true if any of them
// allow the provided action.
func (r *VDIUserRole) Evaluate(action *APIAction) bool {
	for _, rule := range r.Rules {
		if ok := rule.Evaluate(action); ok {
			return true
		}
	}
	return false
}

// IncludesRule returns true if the rules applied to this role are not elevated
// by any of the permissions in the provided rule.
func (r *VDIUserRole) IncludesRule(ruleToCheck Rule, resourceGetter ResourceGetter) bool {
	for _, rule := range r.Rules {
		if ok := rule.IncludesRule(ruleToCheck, resourceGetter); ok {
			return true
		}
	}
	return false
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
