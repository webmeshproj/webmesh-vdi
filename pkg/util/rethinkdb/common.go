package rethinkdb

import "time"

const (
	// The name of the DB in rethinkdb
	kvdiDB = "kvdi"

	// The table for users
	usersTable = "users"
	// The table for user roles
	rolesTable = "roles"
	// The table for user sessions
	userSessionTable = "userSessions"

	// The admin username
	adminUser = "admin"
	// The admin role name
	adminRole = "admin"
	// The anonymous user name
	anonymousUser = "anonymous"
	// The default launch templates role name
	launchTemplateRole = "launch-templates"
)

// allTables is a slice of all the table names in rethinkdb
var allTables = []string{usersTable, rolesTable, userSessionTable}

const (
	// DefaultSessionLength is the session length used for setting expiry
	// times on new user sessions.
	DefaultSessionLength = time.Duration(8) * time.Hour
)
