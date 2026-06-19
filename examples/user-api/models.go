// Package main — models.go
//
// This file defines the data structures (models) used by the user-api example.
// Each struct uses struct tags from GoKit-Lite's validator package so that
// field-level rules are declared directly on the type, keeping validation
// logic close to the data it describes.
package main

// ─── Request Models ───────────────────────────────────────────────────────────

// RegisterRequest represents the JSON body expected by POST /register.
//
// Validation tags applied:
//   - required:"true"    — the field must not be empty
//   - minLength:"2"      — name must have at least 2 characters
//   - maxLength:"50"     — name must have at most 50 characters
//   - email:"true"       — email must be a valid e-mail address
//   - minLength:"6"      — password must have at least 6 characters
type RegisterRequest struct {
	Name     string `json:"name"     required:"true" minLength:"2" maxLength:"50"`
	Email    string `json:"email"    required:"true" email:"true"`
	Password string `json:"password" required:"true" minLength:"6"`
}

// LoginRequest represents the JSON body expected by POST /login.
//
// Validation tags applied:
//   - required:"true" — both fields must be present
//   - email:"true"    — email must follow standard e-mail format
type LoginRequest struct {
	Email    string `json:"email"    required:"true" email:"true"`
	Password string `json:"password" required:"true"`
}

// ─── Response / Domain Models ─────────────────────────────────────────────────

// UserResponse is the safe, public representation of a user returned to
// clients after registration or login. It intentionally omits the password.
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AuthPayload bundles the authenticated user together with the JWT token
// string so both can be returned in a single response envelope.
type AuthPayload struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// ProfileResponse is the public user profile returned by GET /me.
// It is extracted from the validated JWT claims injected by RequireAuth.
type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
