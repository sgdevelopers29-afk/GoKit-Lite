package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/auth"
)

func ExampleGenerateToken() {
	// Initialize global manager settings (typically done once at app startup).
	auth.SetSecret("super-secret-key")
	auth.SetTokenDuration(24 * time.Hour)

	// Create claims for a user
	claims := auth.Claims{
		UserID: "usr_123",
		Email:  "user@example.com",
		Role:   "admin",
	}

	// Generate the token
	token, err := auth.GenerateToken(claims)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}

	// For the example, we just confirm we got a non-empty string.
	// In reality, this token is returned to the client.
	if token != "" {
		fmt.Println("Token generated successfully")
	}

	// Output:
	// Token generated successfully
}

func ExampleValidateToken() {
	auth.SetSecret("super-secret-key")

	// Suppose we have a valid token (generated previously)
	tokenString, _ := auth.GenerateToken(auth.Claims{UserID: "usr_123"})

	// Validate the token string
	parsedClaims, err := auth.ValidateToken(tokenString)
	if err != nil {
		fmt.Println("Error validating token:", err)
		return
	}

	fmt.Printf("Authenticated UserID: %s\n", parsedClaims.UserID)

	// Output:
	// Authenticated UserID: usr_123
}

func ExampleRequireAuth() {
	// 1. Setup Auth Manager
	auth.SetSecret("super-secret-key")

	// 2. Define your protected HTTP handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract claims safely injected by the middleware
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Could not get claims", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Hello, user %s!", claims.UserID)
	})

	// 3. Wrap handler with RequireAuth middleware
	wrappedHandler := auth.RequireAuth(protectedHandler)

	// -- Simulating an HTTP Request --
	token, _ := auth.GenerateToken(auth.Claims{UserID: "456"})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	fmt.Println("Response Code:", rec.Code)
	fmt.Println("Response Body:", rec.Body.String())

	// Output:
	// Response Code: 200
	// Response Body: Hello, user 456!
}
