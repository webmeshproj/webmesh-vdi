package oidc

import (
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// GetUsers should return a list of VDIUsers.
func (a *AuthProvider) GetUsers() ([]*v1alpha1.VDIUser, error) {
	return nil, errors.New("Listing users is not supported when using OIDC authentication")
}

// GetUser should retrieve a single VDIUser.
func (a *AuthProvider) GetUser(username string) (*v1alpha1.VDIUser, error) {
	return nil, errors.New("Retrieving user information is not supported when using OIDC authentication")
}

// CreateUser should handle any logic required to register a new user in kVDI.
func (a *AuthProvider) CreateUser(*v1alpha1.CreateUserRequest) error {
	return errors.New("Creating users is not supported when using OIDC authentication")
}

// UpdateUser should update a VDIUser.
func (a *AuthProvider) UpdateUser(string, *v1alpha1.UpdateUserRequest) error {
	return errors.New("Updating users is not supported when using OIDC authentication")
}

// DeleteUser should remove a VDIUser.
func (a *AuthProvider) DeleteUser(string) error {
	return errors.New("Deleting users is not supported when using OIDC authentication")
}
