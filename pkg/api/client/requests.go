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

package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	rbacv1 "github.com/kvdi/kvdi/apis/rbac/v1"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/errors"

	ktypes "k8s.io/apimachinery/pkg/types"
)

// NamespacedName casts the Kubernetes NamespacedName for ease of use when using this package.
type NamespacedName ktypes.NamespacedName

func (n NamespacedName) String() string { return fmt.Sprintf("%s/%s", n.Namespace, n.Name) }

// Miscellaneous functions

// GetServerVersion will return the version and git commit of the running API server.
func (c *Client) GetServerVersion() (version, gitCommit string, err error) {
	var out map[string]string
	if err = c.do(http.MethodGet, "version", nil, &out); err != nil {
		return
	}
	return out["version"], out["gitCommit"], nil
}

// GetServerConfig returns the VDICluster configuration of the server.
func (c *Client) GetServerConfig() (*appv1.VDIClusterSpec, error) {
	spec := &appv1.VDIClusterSpec{}
	return spec, c.do(http.MethodGet, "config", nil, spec)
}

// GetNamespaces retrieves a list of namespaces the current user has access to.
func (c *Client) GetNamespaces() ([]string, error) {
	var nss []string
	return nss, c.do(http.MethodGet, "namespaces", nil, &nss)
}

// GetServiceAccounts retrieves all the service accounts the current user can use in the given
// namespace.
func (c *Client) GetServiceAccounts(namespace string) ([]string, error) {
	var sas []string
	return sas, c.do(http.MethodGet, fmt.Sprintf("serviceaccounts/%s", namespace), nil, &sas)
}

// WhoAmI retrieves the user details for the currently authenticated account.
func (c *Client) WhoAmI() (*types.VDIUser, error) {
	user := &types.VDIUser{}
	return user, c.do(http.MethodGet, "whoami", nil, user)
}

// Desktop functions

// GetDesktopSessions retrieves the status of currently running desktop sessions in
// kVDI.
func (c *Client) GetDesktopSessions() (*types.DesktopSessionsResponse, error) {
	resp := &types.DesktopSessionsResponse{}
	return resp, c.do(http.MethodGet, "sessions", nil, resp)
}

// CreateDesktopSession creates a new desktop session.
func (c *Client) CreateDesktopSession(opts *types.CreateSessionRequest) (*types.CreateSessionResponse, error) {
	resp := &types.CreateSessionResponse{}
	return resp, c.do(http.MethodPost, "sessions", opts, resp)
}

// DeleteDesktopSession terminates the given desktop session.
func (c *Client) DeleteDesktopSession(nn NamespacedName) error {
	return c.do(http.MethodDelete, fmt.Sprintf("sessions/%s/%s", nn.Namespace, nn.Name), nil, nil)
}

// GetDesktopDisplayProxy returns a ReadWriteCloser proxying the display of the given session.
func (c *Client) GetDesktopDisplayProxy(nn NamespacedName) (io.ReadWriteCloser, error) {
	return c.doWebsocket(fmt.Sprintf("desktops/ws/%s/%s/display", nn.Namespace, nn.Name))
}

// GetDesktopAudioProxy returns a ReadWriteCloser proxying the audio of the given session.
func (c *Client) GetDesktopAudioProxy(nn NamespacedName) (io.ReadWriteCloser, error) {
	return c.doWebsocket(fmt.Sprintf("desktops/ws/%s/%s/audio", nn.Namespace, nn.Name))
}

// StatDesktopFile retrieves stat information for the given path on the desktop.
func (c *Client) StatDesktopFile(nn NamespacedName, path string) (*types.StatDesktopFileResponse, error) {
	resp := &types.StatDesktopFileResponse{}
	return resp, c.do(http.MethodGet, fmt.Sprintf("desktops/fs/%s/%s/stat/%s", nn.Namespace, nn.Name, path), nil, resp)
}

// GetDesktopFile retrieves a ReadCloser containing the contents of the requested file
func (c *Client) GetDesktopFile(nn NamespacedName, path string) (io.ReadCloser, error) {
	resp, err := c.doRaw(http.MethodGet, fmt.Sprintf("desktops/fs/%s/%s/get/%s", nn.Namespace, nn.Name, path), nil)
	if err != nil {
		return nil, err
	}
	if err := errors.CheckAPIError(resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PutDesktopFile uploads a file to the given desktop session.
func (c *Client) PutDesktopFile(nn NamespacedName, name string, contents io.Reader) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, contents); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	r, err := http.NewRequest(http.MethodPut, c.getEndpoint(fmt.Sprintf("desktops/fs/%s/%s/put", nn.Namespace, nn.Name)), &b)
	if err != nil {
		return err
	}
	r.Header.Add("X-Session-Token", c.getAccessToken())
	r.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return errors.CheckAPIError(resp)
}

// VDIRole functions

// GetVDIRoles retrieves the available VDIRoles for kVDI. This is the same as doing
// `kubectl get vdiroles -l "kvdi.io/cluster-ref=kvdi" -o json`.
func (c *Client) GetVDIRoles() ([]*rbacv1.VDIRole, error) {
	resp := make([]*rbacv1.VDIRole, 0)
	return resp, c.do(http.MethodGet, "roles", nil, &resp)
}

// CreateVDIRole creates  a new VDIRole for this cluster.
func (c *Client) CreateVDIRole(req *types.CreateRoleRequest) error {
	return c.do(http.MethodPost, "roles", req, nil)
}

// GetVDIRole retrieves a single VDIRole in kVDI by its name.
func (c *Client) GetVDIRole(name string) (*rbacv1.VDIRole, error) {
	role := &rbacv1.VDIRole{}
	return role, c.do(http.MethodGet, fmt.Sprintf("roles/%s", name), nil, role)
}

// UpdateVDIRole will update a VDIRole. All existing properties are overwritten by those
// in the request, even if nil or unset.
func (c *Client) UpdateVDIRole(name string, req *types.UpdateRoleRequest) error {
	return c.do(http.MethodPut, fmt.Sprintf("roles/%s", name), req, nil)
}

// DeleteVDIRole will delete the given VDIRole.
func (c *Client) DeleteVDIRole(name string) error {
	return c.do(http.MethodDelete, fmt.Sprintf("roles/%s", name), nil, nil)
}

// DesktopTemplate functions

// GetDesktopTemplates returns a list of available DesktopTemplates. This is the same as doing
// `kubectl get desktoptemplates -o json`.
func (c *Client) GetDesktopTemplates() ([]*desktopsv1.Template, error) {
	resp := make([]*desktopsv1.Template, 0)
	return resp, c.do(http.MethodGet, "templates", nil, &resp)
}

// CreateDesktopTemplate creates a new DesktopTemplate for this cluster.
func (c *Client) CreateDesktopTemplate(req *desktopsv1.Template) error {
	return c.do(http.MethodPost, "templates", req, nil)
}

// GetDesktopTemplate retrieves a single DesktopTemplate in kVDI by its name.
func (c *Client) GetDesktopTemplate(name string) (*desktopsv1.Template, error) {
	tmpl := &desktopsv1.Template{}
	return tmpl, c.do(http.MethodGet, fmt.Sprintf("templates/%s", name), nil, tmpl)
}

// UpdateDesktopTemplate will update a DesktopTemplate. Unlike CreateRoleRequest, the
// properties provided in the request are merged into the remote state. So only attributes
// defined in the payload are applied to the remote object.
func (c *Client) UpdateDesktopTemplate(name string, req *desktopsv1.Template) error {
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
func (c *Client) GetVDIUsers() ([]*types.VDIUser, error) {
	resp := make([]*types.VDIUser, 0)
	return resp, c.do(http.MethodGet, "users", nil, &resp)
}

// CreateVDIUser creates a new VDIUser for this cluster, if possible.
func (c *Client) CreateVDIUser(req *types.CreateUserRequest) error {
	return c.do(http.MethodPost, "users", req, nil)
}

// GetVDIUser returns a single VDIUser by name, if possible. VDIUsers are not
// like DesktopTemplates and VDIRoles in that they are not CRDs, and are just used
// as an internal abstraction on the concept of a user.
func (c *Client) GetVDIUser(name string) (*types.VDIUser, error) {
	user := &types.VDIUser{}
	return user, c.do(http.MethodGet, fmt.Sprintf("users/%s", name), nil, user)
}

// UpdateVDIUser will update a VDIUser, if possible. If a password is provided, the
// password is changed for the user. If a list of role names are provided, the user's
// roles are updated to match those provided in the payload.
func (c *Client) UpdateVDIUser(name string, req *types.UpdateUserRequest) error {
	return c.do(http.MethodPut, fmt.Sprintf("users/%s", name), req, nil)
}

// DeleteVDIUser will delete the given VDIUser.
func (c *Client) DeleteVDIUser(name string) error {
	return c.do(http.MethodDelete, fmt.Sprintf("users/%s", name), nil, nil)
}

// TODO: Should MFA management functions be implemented?
