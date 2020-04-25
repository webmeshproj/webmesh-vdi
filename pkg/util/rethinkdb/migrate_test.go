package rethinkdb

import "testing"

func TestMigrate(t *testing.T) {
	mock := NewMock()
	if err := mock.Migrate("password", 1, 1, true); err != nil {
		t.Error(err)
	}
}
