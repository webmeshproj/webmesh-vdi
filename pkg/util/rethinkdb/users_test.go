package rethinkdb

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

func TestGetAllUsers(t *testing.T) {
	mock := NewMock()
	users, err := mock.GetAllUsers()
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
	for idx, name := range []string{adminUser, anonymousUser} {
		if name != users[idx].Name {
			t.Error("Got unexpected user", users[idx])
		}
	}
}

func TestGetUser(t *testing.T) {
	mock := NewMock()
	user, err := mock.GetUser(adminUser)
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
	if user.Name != adminUser {
		t.Error("Got unexpected user name", user.Name)
	}
	if _, err := mock.GetUser(nonExist); err == nil {
		t.Error("Expected error, got nil")
	} else if !errors.IsUserNotFoundError(err) {
		t.Error("Expected a user not found error")
	}
	if _, err := mock.GetUser(errItem); err == nil {
		t.Error("Expected server error, got nil")
	}
}

func TestCreateUser(t *testing.T) {
	mock := NewMock()
	hashFunc = func(string) (string, error) { return testHash, nil }
	if err := mock.CreateUser(&types.User{
		Name: newItem,
	}); err != nil {
		t.Error("Expected no error, got", err)
	}
	if err := mock.CreateUser(&types.User{
		Name: errItem,
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateUser(t *testing.T) {
	mock := NewMock()
	hashFunc = func(string) (string, error) { return testHash, nil }
	if err := mock.UpdateUser(&types.User{
		Name:  newItem,
		Roles: []*types.Role{{Name: newItem}},
	}); err != nil {
		t.Error("Expected no error, got", err)
	}
	if err := mock.UpdateUser(&types.User{
		Name:  errItem,
		Roles: []*types.Role{{Name: errItem}},
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDeleteUser(t *testing.T) {
	mock := NewMock()
	if err := mock.DeleteUser(&types.User{
		Name: newItem,
	}); err != nil {
		t.Error("Expected no error, got", err)
	}
	if err := mock.DeleteUser(&types.User{
		Name: errItem,
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}
