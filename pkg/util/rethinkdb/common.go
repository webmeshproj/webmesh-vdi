package rethinkdb

import "time"

const (
	kvdiDB = "kvdi"

	usersTable           = "users"
	rolesTable           = "roles"
	userSessionTable     = "userSessions"
	desktopSessionsTable = "desktopSessions"

	adminUser = "admin"
	adminRole = "admin"
)

var allTables = []string{usersTable, rolesTable, userSessionTable, desktopSessionsTable}

type RoleGrant string

const (
	GrantAll RoleGrant = "All"
)

var (
	DefaultSessionLength = time.Duration(8) * time.Hour
)
