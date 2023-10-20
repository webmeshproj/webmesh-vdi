/*
Copyright 2020-2023 Avi Zimmerman.

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

// Package webmesh implements an AuthProvider backed by running on a webmesh cluster.
package webmesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	v1 "github.com/kvdi/kvdi/apis/meta/v1"
	"github.com/kvdi/kvdi/pkg/types"
	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/kvdi/kvdi/pkg/util/common"
)

// AuthProvider implements an auth provider that uses a webmesh cluster as the
// authentication backend. Access to groups provided in the claims is supplied
// through annotations on VDIRoles.
type AuthProvider struct {
	metadataURL string
	cluster     *appv1.VDICluster
	client      client.Client
}

// New returns a new AuthProvider.
func New() *AuthProvider {
	return &AuthProvider{}
}

// Reconcile should ensure any k8s resources required for this authentication provider.
func (a *AuthProvider) Reconcile(context.Context, logr.Logger, client.Client, *appv1.VDICluster, string) error {
	return nil
}

// Setup is called when the kVDI app launches and is a chance for the provider
// to setup any resources it needs to serve requests.
func (a *AuthProvider) Setup(cli client.Client, cluster *appv1.VDICluster) error {
	a.metadataURL = cluster.Spec.Auth.WebmeshAuth.MetadataURL
	a.cluster = cluster
	a.client = cli
	return nil
}

// Close is called after temporary uses of the auth provider. It should close
// any open connections and perform cleanup. It should be non-destructive.
func (a *AuthProvider) Close() error {
	return nil
}

// Claims are the claims we expect from the metadata service.
type Claims struct {
	jwt.Claims `json:",inline"`
	Groups     []string `json:"groups"`
}

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *types.LoginRequest) (*types.AuthResult, error) {
	token := req.GetRequest().Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("no Authorization header provided")
	}
	r, err := http.NewRequest("GET", path.Join(a.metadataURL, "id-tokens", "validate"), nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	var claims Claims
	err = json.NewDecoder(resp.Body).Decode(&claims)
	if err != nil {
		return nil, err
	}
	roles, err := a.cluster.GetRoles(a.client)
	if err != nil {
		return nil, err
	}
	return &types.AuthResult{
		User: &types.VDIUser{
			Name: func() string {
				if claims.ID == ":sub" {
					return claims.Subject
				}
				return claims.ID
			}(),
			Roles: func() []*types.VDIUserRole {
				bound := make([]string, 0)
			Roles:
				for _, role := range roles {
					annotations := role.GetAnnotations()
					if annotations == nil {
						continue
					}
					groupstr, ok := annotations[v1.WebmeshGroupRoleAnnotation]
					if !ok {
						continue
					}
					groups := strings.Split(groupstr, v1.AuthGroupSeparator)
					for _, group := range groups {
						if common.StringSliceContains(claims.Groups, group) {
							bound = append(bound, role.GetName())
							continue Roles
						}
					}
				}
				return apiutil.FilterUserRolesByNames(roles, bound)
			}(),
		},
		RefreshNotSupported: true,
	}, nil
}

// GetUsers should return a list of VDIUsers.
func (a *AuthProvider) GetUsers() ([]*types.VDIUser, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetUser should retrieve a single VDIUser.
func (a *AuthProvider) GetUser(string) (*types.VDIUser, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreateUser should handle any logic required to register a new user in kVDI.
func (a *AuthProvider) CreateUser(*types.CreateUserRequest) error {
	return fmt.Errorf("not implemented")
}

// UpdateUser should update a VDIUser.
func (a *AuthProvider) UpdateUser(string, *types.UpdateUserRequest) error {
	return fmt.Errorf("not implemented")
}

// DeleteUser should remove a VDIUser
func (a *AuthProvider) DeleteUser(string) error {
	return fmt.Errorf("not implemented")
}
