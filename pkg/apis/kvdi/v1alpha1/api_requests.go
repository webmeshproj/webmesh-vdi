package v1alpha1

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// API Request/Response types

// LoginRequest represents a request for a session token
type LoginRequest struct {
	// Username
	Username string `json:"username"`
	// Password
	Password string `json:"password"`
}

// SessionResponse represents a response with a new session token
type SessionResponse struct {
	Token     string   `json:"token"`
	ExpiresAt int64    `json:"expiresAt"`
	User      *VDIUser `json:"user"`
}

// CreateUserRequest represents a request to create a new user.
type CreateUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

// Validate the new user request
func (r *CreateUserRequest) Validate() error {
	if r.Username == "" || r.Password == "" {
		return errors.New("'username' and 'password' must be provided in the request")
	}
	if r.Roles == nil || len(r.Roles) == 0 {
		return errors.New("You must assign at least one role to the user")
	}
	if strings.Contains(r.Username, ":") {
		return errors.New("Username cannot contain the ':' character")
	}
	return nil
}

// UpdateUserRequest requests updates to an existing user
type UpdateUserRequest struct {
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

// Validate the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	if r.Password == "" && len(r.Roles) == 0 {
		return errors.New("You must specify either a new password or a list of roles")
	}
	return nil
}

// CreateRoleRequest represents a request for a new role.
type CreateRoleRequest struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

// Validate the CreateRoleRequest
func (r *CreateRoleRequest) Validate() error {
	if r.Name == "" {
		return errors.New("A name is required for the new role")
	}
	for _, rule := range r.Rules {
		if err := validatePatterns(rule.ResourcePatterns); err != nil {
			return err
		}
	}
	return nil
}

func (r *CreateRoleRequest) GetRules() []Rule {
	if r.Rules == nil {
		return []Rule{{
			Verbs:            []Verb{},
			Resources:        []Resource{},
			ResourcePatterns: []string{},
			Namespaces:       []string{},
		}}
	}
	return r.Rules
}

// UpdateRoleRequest requests updates to an existing role.
type UpdateRoleRequest struct {
	Rules []Rule `json:"rules"`
}

func (r *UpdateRoleRequest) GetRules() []Rule {
	if r.Rules == nil {
		return []Rule{{
			Verbs:            []Verb{},
			Resources:        []Resource{},
			ResourcePatterns: []string{},
			Namespaces:       []string{},
		}}
	}
	return r.Rules
}

// Validate the UpdateRoleRequest
func (r *UpdateRoleRequest) Validate() error {
	for _, rule := range r.Rules {
		if err := validatePatterns(rule.ResourcePatterns); err != nil {
			return err
		}
	}
	return nil
}

// validatePatterns takes a list of regexes and returns an error if any of them
// are invalid.
func validatePatterns(patterns []string) error {
	for _, pattern := range patterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("%s is an invalid regex: %s", pattern, err.Error())
		}
	}
	return nil
}

// CreateSessionRequest requests a new desktop session with the givin parameters.
type CreateSessionRequest struct {
	Template  string `json:"template"`
	Namespace string `json:"namespace,omitempty"`
}

// Validate the CreateSessionRequest
func (r *CreateSessionRequest) Validate() error {
	if r.Template == "" {
		return errors.New("A template is required")
	}
	return nil
}

// GetTemplate returns the template for this request
func (r *CreateSessionRequest) GetTemplate() string {
	return r.Template
}

// GetNamespace returns the namspace for this request, or the default namespace
// if not provided.
func (r *CreateSessionRequest) GetNamespace() string {
	if r.Namespace != "" {
		return r.Namespace
	}
	return defaultNamespace
}
