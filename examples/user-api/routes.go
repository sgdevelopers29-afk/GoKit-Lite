// Package main — routes.go
//
// This file wires HTTP routes to handler functions.
// It is intentionally kept separate from main.go (server bootstrap) and
// handlers.go (business logic) so that the routing table is easy to read
// and extend without touching other concerns.
//
// Routing highlights:
//   - Public routes (/health, /register, /login) are registered directly.
//   - Protected routes (/me) are wrapped with auth.RequireAuth so the
//     middleware validates the JWT before the handler is even called.
package main

import (
	"net/http"

	"github.com/sgdevelopers29-afk/GoKit-Lite/auth"
)

// registerRoutes attaches all API routes to the provided ServeMux and returns
// it.  Keeping route registration in its own function makes it straightforward
// to swap in a third-party router (e.g. chi, gorilla/mux) later without
// touching main.go.
func registerRoutes(mux *http.ServeMux) http.Handler {
	// ── Public Routes ──────────────────────────────────────────────────────────
	// These endpoints are reachable without any JWT token.

	// GET /health — liveness probe
	mux.HandleFunc("/health", handleHealth)

	// POST /register — create a new user account
	mux.Handle("/register", methodGuard(http.MethodPost, handleRegister))

	// POST /login — authenticate and receive a JWT
	mux.Handle("/login", methodGuard(http.MethodPost, handleLogin))

	// ── Protected Routes ───────────────────────────────────────────────────────
	// auth.RequireAuth is a standard net/http middleware.
	// It reads the "Authorization: Bearer <token>" header, validates the JWT,
	// and injects the parsed *auth.Claims into r.Context().
	// If the token is missing or invalid it responds 401 and stops the chain.

	// GET /me — return the authenticated user's profile from JWT claims
	mux.Handle("/me", auth.RequireAuth(
		methodGuard(http.MethodGet, handleMe),
	))

	return mux
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// methodGuard returns an http.Handler that only allows the given HTTP method.
// For any other method it responds with 405 Method Not Allowed.
// This keeps handler functions focused on their happy path.
func methodGuard(method string, fn http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w,
				"method not allowed — expected "+method,
				http.StatusMethodNotAllowed)
			return
		}
		fn(w, r)
	})
}
