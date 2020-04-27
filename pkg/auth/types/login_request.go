package types

// LoginRequest represents a request for a session token
type LoginRequest struct {
	// Username
	Username string `json:"username"`
	// Password
	Password string `json:"password"`
}

// Session response represents a response with a new session token
type SessionResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	User      *User  `json:"user"`
}
