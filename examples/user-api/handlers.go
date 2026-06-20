// Package main — handlers.go
//
// This file contains the HTTP handler functions for the user-api example.
// Each handler follows a consistent, three-step pattern that showcases how
// GoKit-Lite's three completed packages work together:
//
//  1. Decode & Validate  — use validator.ValidateAll to collect all field errors.
//  2. Business Logic     — generate a JWT with auth.GenerateToken.
//  3. Respond            — wrap the result with response.Success or response.Error.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sgdevelopers29-afk/GoKit-Lite/auth"
	"github.com/sgdevelopers29-afk/GoKit-Lite/response"
	"github.com/sgdevelopers29-afk/GoKit-Lite/validator"
)

// ─── Helper ───────────────────────────────────────────────────────────────────

// writeJSON encodes v as JSON and writes it to w with the given HTTP status
// code.  All handlers use this helper to keep serialisation in one place.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck // best-effort write
}

// ─── POST /register ───────────────────────────────────────────────────────────

// handleRegister demonstrates the full happy-path and error-path for user
// registration.
//
// Step 1 — Decode the JSON body into a RegisterRequest.
// Step 2 — Validate ALL fields at once using validator.ValidateAll.
//
//	This collects every broken rule in one pass so the client sees all
//	problems in a single response rather than one per request.
//
// Step 3 — Generate a JWT token with auth.GenerateToken.
// Step 4 — Return response.Success with the UserResponse + token.
func handleRegister(w http.ResponseWriter, r *http.Request) {
	// ── Step 1: Decode ────────────────────────────────────────────────────────
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest,
			response.Error("invalid JSON body: "+err.Error()))
		return
	}

	// ── Step 2: Validate ──────────────────────────────────────────────────────
	// ValidateAll runs every rule on every field and gathers all failures.
	// This means a client with three bad fields sees all three errors at once.
	result := validator.ValidateAll(req)
	if !result.Valid {
		// Build a map of fieldName → errorMessage for a clean API response.
		fieldErrors := make(map[string]string, len(result.Errors))
		for _, ve := range result.Errors {
			fieldErrors[ve.Field] = ve.Message
		}
		writeJSON(w, http.StatusUnprocessableEntity,
			response.Error("validation failed"))
		// In a real app you'd embed fieldErrors in a structured envelope.
		// Here we print them to stdout so you can see them while running
		// the example server.
		for field, msg := range fieldErrors {
			fmt.Printf("  [validation] %s: %s\n", field, msg)
		}
		return
	}

	// ── Step 3: Generate JWT ──────────────────────────────────────────────────
	// In a real application you would look the user up in a database here.
	// For the example we fabricate a user ID to keep things self-contained.
	userID := "usr_" + req.Email[:4] // naive demo ID, not production-grade!

	claims := auth.Claims{
		UserID: userID,
		Email:  req.Email,
		Role:   "user",
	}

	token, err := auth.GenerateToken(claims)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			response.Error("could not generate token: "+err.Error()))
		return
	}

	// ── Step 4: Respond ───────────────────────────────────────────────────────
	payload := AuthPayload{
		User: UserResponse{
			ID:    userID,
			Name:  req.Name,
			Email: req.Email,
			Role:  "user",
		},
		Token: token,
	}

	writeJSON(w, http.StatusCreated, response.Success(payload))
}

// ─── POST /login ──────────────────────────────────────────────────────────────

// handleLogin demonstrates the login flow:
//   - Validate the credentials (fail-fast with validator.Validate for speed).
//   - Simulate credential verification.
//   - Issue a JWT and return it with response.Success.
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// ── Step 1: Decode ────────────────────────────────────────────────────────
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest,
			response.Error("invalid JSON body: "+err.Error()))
		return
	}

	// ── Step 2: Validate (fail-fast) ──────────────────────────────────────────
	// For login we use validator.Validate (fail-fast) because we only need
	// to know if the request is structurally sound — we don't need a list of
	// every broken field the way we do during registration.
	if err := validator.Validate(req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity,
			response.Error(err.Error()))
		return
	}

	// ── Step 3: Credential Check ──────────────────────────────────────────────
	// NOTE: In a real application, you would query your database here and
	// compare a hashed password (e.g. bcrypt). This example skips that step
	// to stay focused on showcasing GoKit-Lite APIs.
	const hardcodedPassword = "123456"
	if req.Password != hardcodedPassword {
		// Return 401 without revealing which field is wrong (security best practice).
		writeJSON(w, http.StatusUnauthorized,
			response.Error("invalid email or password"))
		return
	}

	// ── Step 4: Generate JWT ──────────────────────────────────────────────────
	userID := "usr_" + req.Email[:4]
	claims := auth.Claims{
		UserID: userID,
		Email:  req.Email,
		Role:   "user",
	}

	token, err := auth.GenerateToken(claims)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			response.Error("could not generate token: "+err.Error()))
		return
	}

	// ── Step 5: Respond ───────────────────────────────────────────────────────
	payload := AuthPayload{
		User: UserResponse{
			ID:    userID,
			Email: req.Email,
			Role:  "user",
		},
		Token: token,
	}

	writeJSON(w, http.StatusOK, response.Success(payload))
}

// ─── GET /me ──────────────────────────────────────────────────────────────────

// handleMe is a protected endpoint — it is wrapped by auth.RequireAuth in
// routes.go, so this handler is only reached when the token is valid.
//
// It demonstrates how to read the authenticated user's claims from the
// request context using auth.ClaimsFromContext.
func handleMe(w http.ResponseWriter, r *http.Request) {
	// auth.ClaimsFromContext retrieves the *auth.Claims injected by RequireAuth.
	// The second return value is false only if RequireAuth was not applied —
	// which cannot happen on this route, but we guard anyway.
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusInternalServerError,
			response.Error("could not read auth claims from context"))
		return
	}

	profile := ProfileResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}

	writeJSON(w, http.StatusOK, response.Success(profile))
}

// ─── GET /health ──────────────────────────────────────────────────────────────

// handleHealth is a simple liveness probe endpoint.
// It requires no auth and returns an immediate 200 OK — useful for load
// balancers and container orchestrators.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response.Success(map[string]string{
		"status": "ok",
	}))
}
