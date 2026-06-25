# Rate Limiter Package (`ratelimit/`)

## Overview
The `ratelimit` package provides a robust, thread-safe, in-memory rate limiting mechanism for GoKit-Lite. It restricts the number of incoming requests over a given period, acting as a critical defensive layer against brute-force attacks, API abuse, and DDoS attempts.

## Token Bucket Algorithm
This implementation utilizes the highly efficient **Token Bucket** algorithm.
- Each unique key (e.g., an IP address, API key, or User ID) gets its own conceptual "bucket."
- **Burst** specifies the maximum capacity of the bucket. If the bucket is full, requests can be processed immediately up to this burst limit.
- **Rate** specifies how many tokens are added to the bucket every second.
- When a request is made, if the bucket has at least 1 token, the request is allowed and 1 token is consumed. If the bucket is empty, the request is rejected.
- Tokens refill mathematically over time using elapsed duration logic, eliminating the need for expensive background refill goroutines.

## Installation
The `ratelimit` package is part of the GoKit-Lite repository. You can import it into your Go applications using:

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/ratelimit"
```

## Usage Examples

```go
package main

import (
    "fmt"
    "github.com/sgdevelopers29-afk/GoKit-Lite/ratelimit"
)

func main() {
    // Create a new rate limiter allowing 5 requests per second, with a burst capacity of 10
    limiter := ratelimit.New(5, 10)

    // Simulate incoming requests
    ipAddress := "127.0.0.1"

    for i := 1; i <= 15; i++ {
        if limiter.Allow(ipAddress) {
            fmt.Printf("Request %d: Allowed\n", i)
        } else {
            fmt.Printf("Request %d: Rate Limit Exceeded!\n", i)
        }
    }
}
```

## API Reference
```go
// Creates and returns a new rate limiter. 
// `rate` is tokens per second. `burst` is maximum bucket capacity.
func New(rate int, burst int) *Limiter

// Checks if a request for the given key is permitted under the rate limit.
func (l *Limiter) Allow(key string) bool
```

## Performance Notes
- **Lazy Initialization**: Buckets are instantiated strictly on demand. Users who never make a request consume absolutely zero memory.
- **Locking**: State is protected via `sync.RWMutex`, guaranteeing thread safety across concurrent web requests without incurring race conditions.
- **No Background Goroutines**: The Token Bucket relies on lazy-evaluated `time.Now()` calculations on every `Allow()` call. This allows the package to scale infinitely across thousands of keys without leaking background worker goroutines.

## Future Roadmap
- **v2**: Sliding Window algorithm implementation
- **v3**: Middleware integrations (e.g., Gin, Fiber wrappers)
- **v4**: Distributed rate limiting & Redis backend support
- **v5**: Metrics integration (Prometheus)
