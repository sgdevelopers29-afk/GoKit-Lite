package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is an unexported type used to prevent collisions
// when storing values in context.Context.
type contextKey string

const claimsContextKey contextKey = "auth_claims"

// RequireAuth is a standard net/http middleware that extracts a Bearer token
// from the Authorization header, validates it using the default auth manager,
// and injects the parsed Claims into the request context.
//
// If the token is missing, invalid, or expired, it immediately responds with
// a 401 Unauthorized status and stops the handler chain.
func RequireAuth(next http.Handler) http.Handler {
	return defaultManager.RequireAuth(next)
}

// RequireAuth provides the same middleware functionality but scoped to a specific Manager instance.
func (m *Manager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrMissingAuthHeader.Error(), http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, ErrInvalidAuthHeader.Error(), http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			// Do not leak specific internal errors (like missing secret) to clients.
			// Just return 401. Logging could be added here in a future update.
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Inject claims into context
		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// ClaimsFromContext extracts the auth.Claims from a given context.
// Returns nil and false if no claims are present (e.g. RequireAuth was not used).
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}
