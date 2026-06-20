package auth

import "github.com/golang-jwt/jwt/v5"

// Claims defines the standard payload for JWT tokens in GoKit-Lite.
// It embeds jwt.RegisteredClaims to automatically handle standard JWT fields
// such as "exp" (expiration time), "iat" (issued at), and "nbf" (not before).
type Claims struct {
	// UserID uniquely identifies the authenticated user.
	UserID string `json:"user_id,omitempty"`

	// Email is the user's email address.
	Email string `json:"email,omitempty"`

	// Role defines the user's authorization level (e.g., "admin", "user").
	Role string `json:"role,omitempty"`

	// RegisteredClaims provides standard JWT fields (iss, sub, aud, exp, nbf, iat, jti).
	jwt.RegisteredClaims
}
