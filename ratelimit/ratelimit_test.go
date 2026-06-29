package ratelimit_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/ratelimit"
)

func TestNewLimiter(t *testing.T) {
	l := ratelimit.New(5, 10)
	if l == nil {
		t.Fatal("expected new limiter to not be nil")
	}
}

func TestAllowWithinLimit(t *testing.T) {
	l := ratelimit.New(5, 5)
	
	for i := 0; i < 5; i++ {
		if !l.Allow("user1") {
			t.Errorf("expected request %d to be allowed", i)
		}
	}
}

func TestExceedingLimit(t *testing.T) {
	l := ratelimit.New(5, 5)
	
	// Consume all 5 burst tokens
	for i := 0; i < 5; i++ {
		if !l.Allow("user1") {
			t.Fatalf("expected request %d to be allowed", i)
		}
	}

	// 6th request should fail immediately
	if l.Allow("user1") {
		t.Error("expected 6th request to be denied")
	}
}

func TestTokenRefillAfterWaiting(t *testing.T) {
	// Rate is 10 tokens/sec. 1 token = 100ms. Burst = 2.
	l := ratelimit.New(10, 2)

	// Consume burst
	l.Allow("user1")
	l.Allow("user1")

	// 3rd should fail
	if l.Allow("user1") {
		t.Fatal("expected 3rd request to be denied")
	}

	// Wait for 1 token to refill (100ms) plus a tiny buffer
	time.Sleep(110 * time.Millisecond)

	// Should be allowed now
	if !l.Allow("user1") {
		t.Error("expected request to be allowed after waiting for refill")
	}
}

func TestMultipleUsers(t *testing.T) {
	l := ratelimit.New(1, 1)

	// user1 consumes their token
	if !l.Allow("user1") {
		t.Fatal("expected user1 to be allowed")
	}
	if l.Allow("user1") {
		t.Fatal("expected user1 to be denied")
	}

	// user2 should have their own separate bucket and be allowed
	if !l.Allow("user2") {
		t.Error("expected user2 to be allowed")
	}
}

func TestConcurrentRequests(t *testing.T) {
	l := ratelimit.New(10, 100)
	var wg sync.WaitGroup

	successCount := 0
	var mu sync.Mutex

	// Fire 150 concurrent requests for the same user
	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("concurrent_user") {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Since burst is 100, we expect exactly 100 successful requests.
	// (Assuming all requests fire within the same millisecond so no refill occurs during the test).
	if successCount > 100 {
		t.Errorf("expected at most 100 successes, got %d", successCount)
	}
}
