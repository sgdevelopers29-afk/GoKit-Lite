package ratelimit

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// bucket represents the token bucket for a single unique key.
type bucket struct {
	tokens    float64
	lastVisit time.Time
}

// Limiter represents an in-memory rate limiter using the Token Bucket algorithm.
type Limiter struct {
	rate    float64 // tokens added per second
	burst   float64 // maximum tokens the bucket can hold
	buckets map[string]*bucket
	mu      sync.RWMutex
}

// New creates and returns a new rate limiter.
// rate determines how many tokens are added to a bucket per second.
// burst determines the maximum capacity of tokens in a single bucket.
func New(rate int, burst int) *Limiter {
	return &Limiter{
		rate:    float64(rate),
		burst:   float64(burst),
		buckets: make(map[string]*bucket),
	}
}

// Allow checks if a request for the given key is permitted under the rate limit.
// It lazily creates buckets for new keys and mathematically refills tokens based on elapsed time.
// Returns true if the request is allowed, false if the rate limit is exceeded.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, exists := l.buckets[key]
	if !exists {
		// Lazy initialization of a new bucket
		// It starts full of tokens (equal to burst capacity).
		b = &bucket{
			tokens:    l.burst,
			lastVisit: now,
		}
		l.buckets[key] = b
	}

	// Calculate tokens to add based on elapsed time since last visit
	elapsedSeconds := now.Sub(b.lastVisit).Seconds()

	// Refill the bucket
	b.tokens += elapsedSeconds * l.rate

	// Cap the tokens at the maximum burst size
	if b.tokens > l.burst {
		b.tokens = l.burst
	}

	b.lastVisit = now

	// Check if there's at least 1 token available to consume
	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}

	return false
}

// Middleware returns an HTTP middleware that uses the request's RemoteAddr (IP) as the rate limit key.
// Requests that exceed the rate limit receive a 429 Too Many Requests response.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return l.MiddlewareWithKeyFunc(func(r *http.Request) string {
		ip := r.RemoteAddr
		if i := strings.LastIndex(ip, ":"); i != -1 {
			ip = ip[:i] // Strip port
		}
		return ip
	}, next)
}

// MiddlewareWithKeyFunc returns an HTTP middleware that uses a custom function to extract the rate limit key.
func (l *Limiter) MiddlewareWithKeyFunc(keyFunc func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := keyFunc(r)
		if !l.Allow(key) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
