// Package main — main.go
//
// main.go is the entry point for the user-api example.
// It demonstrates the minimal bootstrap required to run a GoKit-Lite-powered
// HTTP server:
//
//  1. Configure the auth package (secret key + token lifetime).
//  2. Register all routes (including protected ones).
//  3. Start the HTTP server.
//
// No third-party router or framework is required — the standard library's
// net/http ServeMux is sufficient for this example.
//
// # Running the server
//
//	cd examples/user-api
//	go run .
//
// The server listens on http://localhost:8080 by default.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/auth"
)

func main() {
	// ── 1. Configure Auth ─────────────────────────────────────────────────────
	//
	// auth.SetSecret configures the HMAC-SHA256 signing key for all JWT
	// operations.  In production this value MUST come from an environment
	// variable or secrets manager — never hard-code it in source code.
	//
	// Example for production:
	//   secret := os.Getenv("JWT_SECRET")
	//   if secret == "" {
	//       log.Fatal("JWT_SECRET environment variable is required")
	//   }
	//   auth.SetSecret(secret)
	auth.SetSecret("gokit-lite-example-secret-change-in-production")

	// auth.SetTokenDuration controls how long newly issued tokens remain valid.
	// 24 hours is the default; we set it explicitly here for clarity.
	auth.SetTokenDuration(24 * time.Hour)

	// ── 2. Register Routes ────────────────────────────────────────────────────
	mux := http.NewServeMux()
	handler := registerRoutes(mux)

	// ── 3. Start Server ───────────────────────────────────────────────────────
	addr := ":8080"

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║         GoKit-Lite — User API Example            ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf( "║  Listening on  http://localhost%s              ║\n", addr)
	fmt.Println("║                                                  ║")
	fmt.Println("║  Routes:                                         ║")
	fmt.Println("║    GET  /health      — liveness probe            ║")
	fmt.Println("║    POST /register    — create account            ║")
	fmt.Println("║    POST /login       — authenticate              ║")
	fmt.Println("║    GET  /me          — profile (auth required)   ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
