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
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"github.com/kvdi/kvdi/pkg/auth/common"
	"github.com/kvdi/kvdi/pkg/types"
)

// AuthProvider implements an auth provider that uses a webmesh cluster as the
// authentication backend. Access to groups provided in the claims is supplied
// through annotations on VDIRoles.
type AuthProvider struct {
	metadataURL string
}

// New returns a new AuthProvider.
func New() common.AuthProvider {
	return &AuthProvider{}
}

// Reconcile should ensure any k8s resources required for this authentication
// provider.
func (a *AuthProvider) Reconcile(context.Context, logr.Logger, client.Client, *appv1.VDICluster, string) error {
	return nil
}

// Setup is called when the kVDI app launches and is a chance for the provider
// to setup any resources it needs to serve requests.
func (a *AuthProvider) Setup(_ client.Client, cluster *appv1.VDICluster) error {
	a.metadataURL = cluster.Spec.Auth.WebmeshAuth.MetadataURL
	return nil
}

// Close is called after temporary uses of the auth provider. It should close
// any open connections and perform cleanup. It should be non-destructive.
func (a *AuthProvider) Close() error {
	return nil
}

// Authenticate is called for API authentication requests. It should generate
// a new JWTClaims object and serve an AuthResult back to the API.
func (a *AuthProvider) Authenticate(req *types.LoginRequest) (*types.AuthResult, error) {
	return nil, fmt.Errorf("not implemented")
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
