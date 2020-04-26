package types

// LoginRequest represents a request for a session token
type LoginRequest struct {
	// Username
	Username string `json:"username"`
	// Password
	Password string `json:"password"`
}
