package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/auth"
)

func main() {
	// 1. Setup the JWT secret and expiration
	auth.SetSecret("super-secret-key-12345")
	auth.SetTokenDuration(1 * time.Hour)

	// 2. Generate a token for a user
	token, err := auth.GenerateToken(auth.Claims{
		UserID: "usr_999",
		Email:  "test@example.com",
		Role:   "admin",
	})
	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}

	fmt.Println("Generated JWT Token:")
	fmt.Println(token)
	fmt.Println()

	// 3. Create a protected HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract claims injected by the middleware
		claims, ok := auth.ClaimsFromContext(r.Context())
		if ok {
			fmt.Fprintf(w, "Hello %s! You have role: %s", claims.Email, claims.Role)
		}
	})

	// 4. Wrap the handler with the RequireAuth middleware
	protectedHandler := auth.RequireAuth(handler)

	fmt.Println("Starting protected server on :8082...")
	fmt.Printf("Try: curl -H \"Authorization: Bearer %s\" http://localhost:8082\n", token)

	// Uncomment to run:
	_ = protectedHandler
	// http.ListenAndServe(":8082", protectedHandler)
}
