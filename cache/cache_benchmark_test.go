package cache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/cache"
)

func BenchmarkCacheSet(b *testing.B) {
	c := cache.New[string, string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", "value")
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := cache.New[string, string]()
	c.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get("key")
	}
}

func BenchmarkCacheDelete(b *testing.B) {
	c := cache.New[string, string]()
	c.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Deleting a missing key evaluates the map lookup and delete performance
		c.Delete("key_non_existent")
	}
}

func BenchmarkCacheSetWithTTL(b *testing.B) {
	c := cache.New[string, string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetWithTTL("key", "value", 5*time.Minute)
	}
}

func BenchmarkCacheParallelGet(b *testing.B) {
	c := cache.New[string, string]()
	c.Set("key", "value")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Get("key")
		}
	})
}

func BenchmarkCacheParallelSet(b *testing.B) {
	c := cache.New[string, string]()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key%d", i), "value")
			i++
		}
	})
}
