package main

import (
	"fmt"
	"net/http"

	"github.com/sgdevelopers29-afk/GoKit-Lite/ratelimit"
)

func main() {
	// Create a rate limiter: 1 token per second, burst capacity of 3
	limiter := ratelimit.New(1, 3)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request Successful!\n"))
	})

	// Wrap the handler with the rate limit middleware
	// This uses the client IP address as the limit key
	protectedHandler := limiter.Middleware(handler)

	fmt.Println("Starting rate-limited server on :8081...")
	fmt.Println("Try hitting it rapidly: curl http://localhost:8081")

	// Uncomment to run:
	_ = protectedHandler
	// http.ListenAndServe(":8081", protectedHandler)
}
