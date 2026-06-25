# Cache Package (`cache/`)

## Overview
The `cache` package provides a generic, thread-safe, in-memory key-value store with Time-To-Live (TTL) support for GoKit-Lite. It leverages Go 1.18+ generics to support arbitrary key and value types without reflection overhead.

## Installation
The `cache` package is part of the GoKit-Lite repository. You can import it into your Go applications using:

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/cache"
```

## Usage

### API Reference
```go
// Creates a new generic cache
func New[K comparable, V any]() *Cache[K, V]

// Adds or updates a key-value pair permanently
func (c *Cache[K, V]) Set(key K, value V)

// Adds or updates a key-value pair with an expiration TTL
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration)

// Retrieves a value by key. Returns false if not found or if expired.
func (c *Cache[K, V]) Get(key K) (V, bool)

// Checks if a key exists and is not expired
func (c *Cache[K, V]) Has(key K) bool

// Deletes a key from the cache
func (c *Cache[K, V]) Delete(key K)

// Returns the total number of items in the cache (including un-swept expired items)
func (c *Cache[K, V]) Size() int

// Removes all entries from the cache
func (c *Cache[K, V]) Clear()

// Starts a background goroutine to clean up expired items periodically
func (c *Cache[K, V]) StartCleanup(interval time.Duration)

// Gracefully stops the cleanup background goroutine
func (c *Cache[K, V]) StopCleanup()
```

### Example: Basic and TTL caching

```go
package main

import (
    "fmt"
    "time"
    "github.com/sgdevelopers29-afk/GoKit-Lite/cache"
)

func main() {
    // Create a new thread-safe cache
    userCache := cache.New[string, string]()
    
    // Optional: Start a cleanup worker to sweep expired entries every 5 minutes
    userCache.StartCleanup(5 * time.Minute)
    defer userCache.StopCleanup()

    // Set a permanent value
    userCache.Set("admin", "Shreyas")

    // Set a value with TTL
    userCache.SetWithTTL("session_token", "abc123xyz", 30*time.Second)

    // Retrieve values
    if value, ok := userCache.Get("admin"); ok {
        fmt.Println("Admin:", value)
    }

    if value, ok := userCache.Get("session_token"); ok {
        fmt.Println("Token:", value)
    }

    // Wait for expiration
    time.Sleep(31 * time.Second)
    if _, ok := userCache.Get("session_token"); !ok {
        fmt.Println("Token has successfully expired.")
    }
}
```

## Thread Safety Guarantee
All methods on `Cache[K, V]` are safe for concurrent use across multiple goroutines. Read-heavy workloads benefit from `sync.RWMutex`, which permits concurrent reads via `Get` while ensuring exclusive locks during `Set`, `Delete`, and background sweeps.

## Future Roadmap

- **v3**: LRU/LFU cache policies
- **v4**: Distributed cache interface / Redis backend integration
- **v5**: Persistence to disk
