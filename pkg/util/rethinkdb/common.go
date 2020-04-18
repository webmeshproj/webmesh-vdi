package rethinkdb

import "time"

const (
	kvdiDB = "kvdi"

	usersTable           = "users"
	rolesTable           = "roles"
	userSessionTable     = "userSessions"
	desktopSessionsTable = "desktopSessions"

	adminUser          = "admin"
	adminRole          = "admin"
	anonymousUser      = "anonymous"
	launchTemplateRole = "launch-templates"
)

var allTables = []string{usersTable, rolesTable, userSessionTable, desktopSessionsTable}

const (
	DefaultSessionLength = time.Duration(8) * time.Hour
)
