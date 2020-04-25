package rethinkdb

import (
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

func TestGetAllRoles(t *testing.T) {
	mock := NewMock()
	roles, err := mock.GetAllRoles()
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
	for idx, name := range []string{adminRole, launchTemplateRole} {
		if name != roles[idx].Name {
			t.Error("Got unexpected role", roles[idx])
		}
	}
}

func TestGetRole(t *testing.T) {
	mock := NewMock()
	role, err := mock.GetRole(adminRole)
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
	if role.Name != adminRole {
		t.Error("Got unexpected role name", role.Name)
	}
	if _, err := mock.GetRole(nonExist); err == nil {
		t.Error("Expected error, got nil")
	} else if !errors.IsRoleNotFoundError(err) {
		t.Error("Expected a role not found error")
	}
	if _, err := mock.GetRole(errItem); err == nil {
		t.Error("Expected server error, got nil")
	}
}

func TestCreateRole(t *testing.T) {
	mock := NewMock()
	if err := mock.CreateRole(&types.Role{
		Name: newItem,
	}); err != nil {
		t.Error("Expected no error, got", err)
	}
	if err := mock.CreateRole(&types.Role{
		Name: errItem,
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateRole(t *testing.T) {
	mock := NewMock()
	if err := mock.UpdateRole(&types.Role{
		Name: newItem,
	}); err != nil {
		t.Error("Expected no error, got", err)
	}
	if err := mock.UpdateRole(&types.Role{
		Name: errItem,
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDeleteRole(t *testing.T) {
	mock := NewMock()
	if err := mock.DeleteRole(&types.Role{
		Name: newItem,
	}); err != nil {
		t.Error("Expected no error, got", err)
	}
	if err := mock.DeleteRole(&types.Role{
		Name: errItem,
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}
