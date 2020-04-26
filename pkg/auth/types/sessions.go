package types

import "time"

// UserSession represents a user API session, usually initiated from the
// app frontend by logging in, however can also be used for programatic access
// if desired.
type UserSession struct {
	Token     string    `rethinkdb:"id" json:"token"`
	ExpiresAt time.Time `rethinkdb:"expires_at" json:"expiresAt"`
	User      *User     `rethinkdb:"user_id,reference" rethinkdb_ref:"id" json:"user"`
}
