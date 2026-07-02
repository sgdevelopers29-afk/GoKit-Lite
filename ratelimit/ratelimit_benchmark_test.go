package ratelimit_test

import (
	"fmt"
	"testing"

	"github.com/sgdevelopers29-afk/GoKit-Lite/ratelimit"
)

func BenchmarkAllow(b *testing.B) {
	l := ratelimit.New(1000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Allow("user")
	}
}

func BenchmarkAllowParallel(b *testing.B) {
	l := ratelimit.New(1000000, 1000000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Allow("user")
		}
	})
}

func BenchmarkMultipleClients(b *testing.B) {
	l := ratelimit.New(100, 100)
	// Create a large number of distinct clients
	clients := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		clients[i] = fmt.Sprintf("user_%d", i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Round robin through clients to test map access dispersion
			l.Allow(clients[i%1000])
			i++
		}
	})
}

func BenchmarkTokenRefill(b *testing.B) {
	// Refill computation overhead
	l := ratelimit.New(10, 10)
	l.Allow("user") // Force lazy init
	b.ResetTimer()

	// Running this in a tight loop tests the computational overhead of the token math
	for i := 0; i < b.N; i++ {
		l.Allow("user")
	}
}
