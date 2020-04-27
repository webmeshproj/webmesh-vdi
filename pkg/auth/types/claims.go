package types

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type JWTClaims struct {
	User *User `json:"user"`
	jwt.StandardClaims
}

const (
	// DefaultSessionLength is the session length used for setting expiry
	// times on new user sessions.
	DefaultSessionLength = time.Duration(8) * time.Hour
)
