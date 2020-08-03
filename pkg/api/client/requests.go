package client

import (
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

// Miscellaneous functions

// GetServerConfig returns the VDICluster configuration of the server.
func (c *Client) GetServerConfig() (*v1alpha1.VDIClusterSpec, error) {
	spec := &v1alpha1.VDIClusterSpec{}
	return spec, c.do(http.MethodGet, "config", nil, spec)
}

// GetNamespaces retrieves a list of namespaces the current user has access to.
func (c *Client) GetNamespaces() ([]string, error) {
	nss := make([]string, 0)
	return nss, c.do(http.MethodGet, "namespaces", nil, &nss)
}

// WhoAmI retrieves the user details for the currently authenticated account.
func (c *Client) WhoAmI() (*v1.VDIUser, error) {
	user := &v1.VDIUser{}
	return user, c.do(http.MethodGet, "whoami", nil, user)
}

// Desktop functions

// GetDesktopSessions retrieves the status of currently running desktop sessions in
// kVDI.
func (c *Client) GetDesktopSessions() (*v1.DesktopSessionsResponse, error) {
	resp := &v1.DesktopSessionsResponse{}
	return resp, c.do(http.MethodGet, "sessions", nil, resp)
}

// TODO: Should Create,Use,Delete desktop sessions be implemented?

// VDIRole functions

// GetVDIRoles retrieves the available VDIRoles for kVDI. This is the same as doing
// `kubectl get vdiroles -l "kvdi.io/cluster-ref=kvdi" -o json`.
func (c *Client) GetVDIRoles() ([]*v1alpha1.VDIRole, error) {
	resp := make([]*v1alpha1.VDIRole, 0)
	return resp, c.do(http.MethodGet, "roles", nil, &resp)
}

// CreateVDIRole creates  a new VDIRole for this cluster.
func (c *Client) CreateVDIRole(req *v1.CreateRoleRequest) error {
	return c.do(http.MethodPost, "roles", req, nil)
}

// GetVDIRole retrieves a single VDIRole in kVDI by its name.
func (c *Client) GetVDIRole(name string) (*v1alpha1.VDIRole, error) {
	role := &v1alpha1.VDIRole{}
	return role, c.do(http.MethodGet, fmt.Sprintf("roles/%s", name), nil, role)
}

// UpdateVDIRole will update a VDIRole. All existing properties are overwritten by those
// in the request, even if nil or unset.
func (c *Client) UpdateVDIRole(name string, req *v1.UpdateRoleRequest) error {
	return c.do(http.MethodPut, fmt.Sprintf("roles/%s", name), req, nil)
}

// DeleteVDIRole will delete the given VDIRole.
func (c *Client) DeleteVDIRole(name string) error {
	return c.do(http.MethodDelete, fmt.Sprintf("roles/%s", name), nil, nil)
}

// DesktopTemplate functions

// GetDesktopTemplates returns a list of available DesktopTemplates. This is the same as doing
// `kubectl get desktoptemplates -o json`.
func (c *Client) GetDesktopTemplates() ([]*v1alpha1.DesktopTemplate, error) {
	resp := make([]*v1alpha1.DesktopTemplate, 0)
	return resp, c.do(http.MethodGet, "templates", nil, &resp)
}

// CreateDesktopTemplate creates a new DesktopTemplate for this cluster.
func (c *Client) CreateDesktopTemplate(req *v1alpha1.DesktopTemplate) error {
	return c.do(http.MethodPost, "templates", req, nil)
}

// GetDesktopTemplate retrieves a single DesktopTemplate in kVDI by its name.
func (c *Client) GetDesktopTemplate(name string) (*v1alpha1.DesktopTemplate, error) {
	tmpl := &v1alpha1.DesktopTemplate{}
	return tmpl, c.do(http.MethodGet, fmt.Sprintf("templates/%s", name), nil, tmpl)
}

// UpdateDesktopTemplate will update a DesktopTemplate. Unlike CreateRoleRequest, the
// properties provided in the request are merged into the remote state. So only attributes
// defined in the payload are applied to the remote object.
func (c *Client) UpdateDesktopTemplate(name string, req *v1alpha1.DesktopTemplate) error {
	return c.do(http.MethodPut, fmt.Sprintf("templates/%s", name), req, nil)
}

// DeleteDesktopTemplate will delete the given DesktopTemplate.
func (c *Client) DeleteDesktopTemplate(name string) error {
	return c.do(http.MethodDelete, fmt.Sprintf("templates/%s", name), nil, nil)
}

// VDIUser functions

// GetVDIUsers returns a list of available VDIUsers, if possible. VDIUsers are not
// like DesktopTemplates and VDIRoles in that they are not CRDs, and are just used
// as an internal abstraction on the concept of a user.
func (c *Client) GetVDIUsers() ([]*v1.VDIUser, error) {
	resp := make([]*v1.VDIUser, 0)
	return resp, c.do(http.MethodGet, "users", nil, &resp)
}

// CreateVDIUser creates a new VDIUser for this cluster, if possible.
func (c *Client) CreateVDIUser(req *v1.CreateUserRequest) error {
	return c.do(http.MethodPost, "users", req, nil)
}

// GetVDIUser returns a single VDIUser by name, if possible. VDIUsers are not
// like DesktopTemplates and VDIRoles in that they are not CRDs, and are just used
// as an internal abstraction on the concept of a user.
func (c *Client) GetVDIUser(name string) (*v1.VDIUser, error) {
	user := &v1.VDIUser{}
	return user, c.do(http.MethodGet, fmt.Sprintf("users/%s", name), nil, user)
}

// UpdateVDIUser will update a VDIUse, if possibler. If a password is provided, the
// password is changed for the user. If a list of role names are provided, the user's
// roles are updated to match those provided in the payload.
func (c *Client) UpdateVDIUser(name string, req *v1.UpdateUserRequest) error {
	return c.do(http.MethodPut, fmt.Sprintf("users/%s", name), req, nil)
}

// DeleteVDIUser will delete the given VDIUser.
func (c *Client) DeleteVDIUser(name string) error {
	return c.do(http.MethodDelete, fmt.Sprintf("users/%s", name), nil, nil)
}

// TODO: Should MFA management functions be implemented?
