# Cache Package (`cache/`)

## Overview
The `cache` package provides a generic, lightweight, in-memory key-value store for GoKit-Lite. It is designed for simplicity and speed, leveraging Go generics to support arbitrary key and value types without reflection or type assertions. 

**Note:** v1 of this package is *not* thread-safe and does not support automatic TTL-based evictions.

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

// Adds or updates a key-value pair
func (c *Cache[K, V]) Set(key K, value V)

// Retrieves a value by key. Returns the value and a boolean indicating if it was found
func (c *Cache[K, V]) Get(key K) (V, bool)

// Deletes a key from the cache
func (c *Cache[K, V]) Delete(key K)

// Returns the total number of items in the cache
func (c *Cache[K, V]) Size() int

// Removes all entries from the cache
func (c *Cache[K, V]) Clear()
```

### Examples

```go
package main

import (
    "fmt"
    "github.com/sgdevelopers29-afk/GoKit-Lite/cache"
)

func main() {
    // Create a new cache with string keys and string values
    userCache := cache.New[string, string]()

    // Set a value
    userCache.Set("name", "Shreyas")

    // Get a value
    if value, ok := userCache.Get("name"); ok {
        fmt.Println("Found:", value)
    }

    // Check size
    fmt.Println("Cache Size:", userCache.Size()) // Output: 1

    // Delete a value
    userCache.Delete("name")
}
```

## Future Roadmap

- **v2**: Add `sync.RWMutex` support for concurrent access protection.
- **v3**: Implement TTL cache and automatic cleanup goroutines.
- **v4**: Implement LRU cache and metrics support.
