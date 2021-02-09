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

package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/api/client"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

// mustNewTestAPI creates and starts a new HTTP server connected to the
// API routes. It is initialized with a fake client loaded with core resources
// required for runtime.
func mustNewTestAPI(t *testing.T) (*http.Server, *client.Opts) {
	t.Helper()
	srvr, addr, passw, err := NewTestAPI()
	if err != nil {
		t.Fatal(err)
	}
	return srvr, &client.Opts{
		URL:      addr,
		Username: "admin",
		Password: passw,
	}
}

// mustNewClientWithClose creates a test server, connects a client to it,
// and defines a close function to stop both cleanly.
// The client and the function are returned,
func mustNewClientWithClose(t *testing.T) (*client.Client, func()) {
	t.Helper()
	srvr, opts := mustNewTestAPI(t)
	cl, err := client.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	return cl, func() {
		cl.Close()
		srvr.Close()
	}
}

// TestUsers tests user related operations.
func TestUsers(t *testing.T) {
	cl, close := mustNewClientWithClose(t)
	defer close()

	// check that users just returns an admin user
	users, err := cl.GetVDIUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].GetName() != "admin" {
		t.Error("Expected one admin user, got", users)
	}

	// check that we can query the admin user separately
	if _, err := cl.GetVDIUser("admin"); err != nil {
		t.Error("Expected to be able to get admin user, got:", err)
	}

	// Check that we can't create a user without a password
	if err := cl.CreateVDIUser(&v1.CreateUserRequest{
		Username: "test-user",
		Roles:    []string{"test-cluster-admin"},
	}); err == nil {
		t.Error("Expected to not be able to create user with no password, got nil error")
	} else if !strings.Contains(err.Error(), "'password' must be provided") {
		t.Error("Expected error related to unassigned roles, got:", err)
	}

	// Check that we can't create a user without roles
	if err := cl.CreateVDIUser(&v1.CreateUserRequest{
		Username: "test-user",
		Password: "test-password",
	}); err == nil {
		t.Error("Expected to not be able to create user with no roles, got nil error")
	} else if !strings.Contains(err.Error(), "assign at least one role") {
		t.Error("Expected error related to unassigned roles, got:", err)
	}

	// Check that we can create a user
	if err := cl.CreateVDIUser(&v1.CreateUserRequest{
		Username: "test-user",
		Password: "test-password",
		Roles:    []string{"test-cluster-admin"},
	}); err != nil {
		t.Fatal("Unable to create test user:", err)
	}

	// Check that we now have two users
	users, err = cl.GetVDIUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Error("Expected two users, got", users)
	}

	// Retrieve the user to do some updates to it
	newUser, err := cl.GetVDIUser("test-user")
	if err != nil {
		t.Fatal(err)
	}

	// do some validations on the created user
	if len(newUser.Roles) != 1 {
		t.Error("Expected new user to have one role, got:", newUser.Roles)
	} else if newUser.Roles[0].GetName() != "test-cluster-admin" {
		t.Error("Expected new user to have cluster admin role, got:", newUser.Roles[0].GetName())
	}

	// change the users password, authentication tests will verify efficacy of this
	if err := cl.UpdateVDIUser("test-user", &v1.UpdateUserRequest{
		Password: "new-password",
	}); err != nil {
		t.Fatal(err)
	}

	// change the users role
	if err := cl.UpdateVDIUser("test-user", &v1.UpdateUserRequest{
		Roles: []string{"test-cluster-launch-templates"},
	}); err != nil {
		t.Fatal(err)
	}

	// Retrieve the user and see if the role changed
	newUser, err = cl.GetVDIUser("test-user")
	if err != nil {
		t.Fatal(err)
	}
	if len(newUser.Roles) != 1 {
		t.Error("Expected new user to have one role, got:", newUser.Roles)
	} else if newUser.Roles[0].GetName() != "test-cluster-launch-templates" {
		t.Error("Expected new user to have cluster launch-templates role, got:", newUser.Roles[0].GetName())
	}

	// Delete the user
	if err := cl.DeleteVDIUser("test-user"); err != nil {
		t.Fatal(err)
	}

	// check that users just returns an admin user
	users, err = cl.GetVDIUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].GetName() != "admin" {
		t.Error("Expected one admin user, got", users)
	}

	// user shouldn't be found
	if _, err = cl.GetVDIUser("test-user"); err == nil {
		t.Error("Expected error for retrieving deleted user, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Error("Expected user not found error, got:", err)
	}

	// same for update
	if err = cl.UpdateVDIUser("test-user", &v1.UpdateUserRequest{Password: "blah"}); err == nil {
		t.Error("Expected error for retrieving deleted user, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Error("Expected user not found error, got:", err)
	}

}
