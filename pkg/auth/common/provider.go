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

package common

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthProvider defines an interface for handling login attempts. Currently
// only local auth (using the secrets backend) is supported, however other integrations
// such as LDAP or OAuth can implement this interface.
type AuthProvider interface {
	// Reconcile should ensure any k8s resources required for this authentication
	// provider.
	Reconcile(logr.Logger, client.Client, *v1alpha1.VDICluster, string) error
	// Setup is called when the kVDI app launches and is a chance for the provider
	// to setup any resources it needs to serve requests.
	Setup(client.Client, *v1alpha1.VDICluster) error
	// Close is called after temporary uses of the auth provider. It should close
	// any open connections and perform cleanup. It should be non-destructive.
	Close() error

	// API helper methods
	// Not all providers will be able to implement all of these methods. When
	// they can't they should serve a concise error message as to why.

	// Authenticate is called for API authentication requests. It should generate
	// a new JWTClaims object and serve an AuthResult back to the API.
	Authenticate(*v1.LoginRequest) (*v1.AuthResult, error)
	// GetUsers should return a list of VDIUsers.
	GetUsers() ([]*v1.VDIUser, error)
	// GetUser should retrieve a single VDIUser.
	GetUser(string) (*v1.VDIUser, error)
	// CreateUser should handle any logic required to register a new user in kVDI.
	CreateUser(*v1.CreateUserRequest) error
	// UpdateUser should update a VDIUser.
	UpdateUser(string, *v1.UpdateUserRequest) error
	// DeleteUser should remove a VDIUser
	DeleteUser(string) error
}
