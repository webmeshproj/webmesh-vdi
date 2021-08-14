/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package types

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"

	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
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

// VDIUserRole represents a VDIRole, but only with the data that is to be
// embedded in the JWT. Primarily, leaving out useless metadata that will inflate
// the token.
type VDIUserRole struct {
	// The name of the role, this must match the VDIRole from which this object
	// derives.
	Name string `json:"name"`
	// The rules for this role.
	Rules []rbacv1.Rule `json:"rules"`
}

// GetName returns the name of the role
func (r *VDIUserRole) GetName() string { return r.Name }

// APIAction represents an API action to evaluate against a user's roles.
type APIAction struct {
	// The verb type of the action
	Verb rbacv1.Verb `json:"verb"`
	// The resource type of the action
	ResourceType rbacv1.Resource `json:"resourceType"`
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
