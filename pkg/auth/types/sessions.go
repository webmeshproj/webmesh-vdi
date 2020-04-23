package types

import "time"

type UserSession struct {
	Token     string    `rethinkdb:"id" json:"token"`
	ExpiresAt time.Time `rethinkdb:"expires_at" json:"expiresAt"`
	User      *User     `rethinkdb:"user_id,reference" rethinkdb_ref:"id" json:"user"`
}
