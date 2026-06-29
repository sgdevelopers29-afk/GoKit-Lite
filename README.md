<p align="center">
  <h1 align="center">GoKit-Lite</h1>
  <p align="center">
    A lightweight, modular toolkit for building production-grade Go backends.
    <br />
    <a href="docs/getting-started.md"><strong>Getting Started »</strong></a>
    &nbsp;&middot;&nbsp;
    <a href="docs/architecture.md">Architecture</a>
    &nbsp;&middot;&nbsp;
    <a href="examples/">Examples</a>
    &nbsp;&middot;&nbsp;
    <a href="https://github.com/sgdevelopers29-afk/GoKit-Lite/issues">Report Bug</a>
  </p>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/sgdevelopers29-afk/GoKit-Lite"><img src="https://pkg.go.dev/badge/github.com/sgdevelopers29-afk/GoKit-Lite.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/sgdevelopers29-afk/GoKit-Lite"><img src="https://goreportcard.com/badge/github.com/sgdevelopers29-afk/GoKit-Lite" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/sgdevelopers29-afk/GoKit-Lite/pulls"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg" alt="PRs Welcome"></a>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go Version">
</p>

---

## Why GoKit-Lite?

Every Go backend project eventually needs the same building blocks — standardized API responses, input validation, JWT authentication, caching, rate limiting, and monitoring. Most teams either cobble these together from scattered packages with conflicting conventions or rewrite them from scratch each time.

GoKit-Lite solves this by packaging these recurring patterns into a single, cohesive toolkit:

- **Stop reinventing infrastructure** — battle-tested utilities out of the box.
- **Consistent API contracts** — all packages follow the same design patterns and documentation standards.
- **Gradual adoption** — import one package today, add more as you need them. No all-or-nothing commitment.
- **No vendor lock-in** — every package works with `net/http` and any router of your choice.

---

## Design Principles

| Principle | What It Means |
|-----------|---------------|
| **Standard library first** | Only one external dependency (`golang-jwt/jwt` for `auth`). Everything else is pure stdlib. |
| **Framework agnostic** | Works with `net/http`, Gin, Echo, Chi, Fiber, or any Go HTTP framework. |
| **Modular by design** | Each package is fully independent — import only what you need. |
| **Production-ready** | Thread-safe, well-tested, benchmarked, and documented. |
| **Easy to extend** | Register custom validators, create scoped auth managers, or build named monitors. |
| **Minimal API surface** | Small, focused interfaces that are easy to learn and hard to misuse. |

---

## Compatibility

GoKit-Lite is **framework-independent**. All packages use standard Go interfaces (`http.Handler`, `context.Context`, `error`) and work seamlessly with:

| Framework | Compatible | Notes |
|-----------|:----------:|-------|
| `net/http` | ✅ | Native support — all middleware and handlers use `http.Handler` |
| [Chi](https://github.com/go-chi/chi) | ✅ | `auth.RequireAuth` works directly as Chi middleware |
| [Gin](https://github.com/gin-gonic/gin) | ✅ | Wrap with `gin.WrapH()` for middleware, use packages directly in handlers |
| [Echo](https://github.com/labstack/echo) | ✅ | Wrap with `echo.WrapMiddleware()` for middleware |
| [Fiber](https://github.com/gofiber/fiber) | ✅ | Use adaptor package (`fibadaptor`) to bridge `http.Handler` |

> **Note:** GoKit-Lite does not import or depend on any of these frameworks. It produces and consumes standard library types only.

---

## Features

| Package | Description | Status | Docs |
|---------|-------------|:------:|------|
| [`response`](response/) | Standardized JSON API envelope with `Success()` and `Error()` helpers | ✅ Stable | [response.md](docs/response.md) |
| [`validator`](validator/) | Tag-based struct validation — 9 built-in rules, custom validators, recursive nested support | ✅ Stable | [validator.md](docs/validator.md) |
| [`auth`](auth/) | JWT generation, validation, and `net/http` middleware using HMAC-SHA256 | ✅ Stable | [auth.md](docs/auth.md) |
| [`config`](config/) | Environment variable management with `.env` file loading | ✅ Stable | [config.md](docs/config.md) |
| [`cache`](cache/) | Generic, thread-safe in-memory cache with TTL and background cleanup | ✅ Stable | [cache.md](docs/cache.md) |
| [`logger`](logger/) | Structured JSON logging with `INFO` and `ERROR` levels | ✅ Stable | [logger.md](docs/logger.md) |
| [`monitor`](monitor/) | Lock-free request metrics — counts, success rate, and latency tracking | ✅ Stable | [monitor.md](docs/monitor.md) |
| [`ratelimit`](ratelimit/) | Token-bucket rate limiter with per-key tracking | ✅ Stable | [ratelimit.md](docs/ratelimit.md) |

---

## Architecture

### Package Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        Your Application                              │
│                                                                      │
│  ┌───────────┐ ┌───────────┐ ┌──────────┐ ┌────────┐ ┌───────────┐ │
│  │ response  │ │ validator │ │   auth   │ │ config │ │   cache   │ │
│  │           │ │           │ │          │ │        │ │           │ │
│  │ Success() │ │ Validate()│ │ Generate │ │ Load() │ │ Get/Set   │ │
│  │ Error()   │ │ ValidAll()│ │ Validate │ │ Get()  │ │ TTL       │ │
│  │           │ │ Register()│ │ Require  │ │        │ │ Cleanup   │ │
│  └───────────┘ └───────────┘ └──────────┘ └────────┘ └───────────┘ │
│                                                                      │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐                         │
│  │  logger   │ │  monitor  │ │ ratelimit │                         │
│  │           │ │           │ │           │                         │
│  │ Info()    │ │ Record*() │ │ Allow()   │                         │
│  │ Error()   │ │ GetStats()│ │           │                         │
│  └───────────┘ └───────────┘ └───────────┘                         │
│                                                                      │
│  Each package is independent — no inter-package dependencies.        │
│  Only auth depends on golang-jwt/jwt. Everything else is stdlib.     │
└──────────────────────────────────────────────────────────────────────┘
```

### Typical Request Flow

```
Client Request
     │
     ▼
┌─────────────┐
│  ratelimit  │──── Too many requests? → 429 Response
│  Allow()    │
└──────┬──────┘
       │ allowed
       ▼
┌─────────────┐
│    auth     │──── Invalid/missing token? → 401 Response
│ RequireAuth │
└──────┬──────┘
       │ authenticated
       ▼
┌─────────────┐
│  validator  │──── Validation failed? → 422 Response
│ Validate()  │
└──────┬──────┘
       │ valid
       ▼
┌─────────────┐     ┌─────────────┐
│  Business   │────►│   cache     │  (optional: cache lookups)
│   Logic     │     │ Get/Set     │
└──────┬──────┘     └─────────────┘
       │
       ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  response   │     │  monitor    │     │   logger    │
│ Success()   │     │ Record*()   │     │ Info/Error  │
│  Error()    │     │ GetStats()  │     │             │
└──────┬──────┘     └─────────────┘     └─────────────┘
       │
       ▼
  JSON Response
```

> See [docs/architecture.md](docs/architecture.md) for detailed request lifecycle diagrams with code examples.

---

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

**Requirements:** Go 1.25 or later.

---

## Quick Start

### Standardized API Responses

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/response"

// Success — returns {"success":true,"message":"success","data":{...}}
resp := response.Success(map[string]string{"name": "Ganesh"})

// Error — returns {"success":false,"message":"User not found"}
errResp := response.Error("User not found")
```

### Input Validation

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/validator"

type SignupRequest struct {
    Name     string `required:"true"`
    Email    string `required:"true" email:"true"`
    Password string `required:"true" minLength:"8"`
    Role     string `oneOf:"admin,user"`
}

if err := validator.Validate(req); err != nil {
    // err contains the first validation failure
}

// Or collect all errors at once:
result := validator.ValidateAll(req)
if !result.Valid {
    for _, e := range result.Errors {
        fmt.Printf("%s: %s\n", e.Field, e.Message)
    }
}
```

### JWT Authentication

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/auth"

// Configure
auth.SetSecret(os.Getenv("JWT_SECRET"))
auth.SetTokenDuration(24 * time.Hour)

// Generate
token, err := auth.GenerateToken(auth.Claims{
    UserID: "usr_001",
    Email:  "alice@example.com",
    Role:   "admin",
})

// Validate
claims, err := auth.ValidateToken(token)

// Middleware — protect routes
mux.Handle("/api/me", auth.RequireAuth(myHandler))

// Extract claims in handler
claims, ok := auth.ClaimsFromContext(r.Context())
```

### Configuration

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/config"

// Load .env file
config.Load(".env")

// Read environment variables
port := config.Get("PORT")
```

### Caching

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/cache"

c := cache.New[string, User]()

c.SetWithTTL("user:001", user, 5*time.Minute)

if u, ok := c.Get("user:001"); ok {
    // cache hit
}

// Background cleanup of expired entries
c.StartCleanup(1 * time.Minute)
defer c.StopCleanup()
```

### Structured Logging

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/logger"

logger.Info("Server started on port 8080")
logger.Error("Database connection failed")
// Output: {"level":"INFO","message":"Server started on port 8080","timestamp":"2026-01-01T00:00:00Z"}
```

### Monitoring

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/monitor"

start := time.Now()
monitor.RecordRequest()

// ... handle request ...

monitor.RecordSuccess()
monitor.RecordLatency(time.Since(start))

stats := monitor.GetStats()
fmt.Printf("Requests: %d, Success Rate: %.2f%%\n", stats.Requests, stats.SuccessRate)
```

### Rate Limiting

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/ratelimit"

limiter := ratelimit.New(10, 20) // 10 tokens/sec, burst of 20

if !limiter.Allow(clientIP) {
    http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
    return
}
```

---

## Integrated Example: Response + Validator + Auth

This example demonstrates how multiple GoKit-Lite packages compose together in a single HTTP handler:

```go
package main

import (
    "encoding/json"
    "net/http"
    "os"
    "time"

    "github.com/sgdevelopers29-afk/GoKit-Lite/auth"
    "github.com/sgdevelopers29-afk/GoKit-Lite/response"
    "github.com/sgdevelopers29-afk/GoKit-Lite/validator"
)

type LoginRequest struct {
    Email    string `json:"email"    required:"true" email:"true"`
    Password string `json:"password" required:"true" minLength:"8"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // 1. Decode request
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(response.Error("invalid request body"))
        return
    }

    // 2. Validate input
    if err := validator.Validate(req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(response.Error(err.Error()))
        return
    }

    // 3. Generate JWT
    token, err := auth.GenerateToken(auth.Claims{
        UserID: "usr_001",
        Email:  req.Email,
        Role:   "user",
    })
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(response.Error("authentication failed"))
        return
    }

    // 4. Return standardized response
    json.NewEncoder(w).Encode(response.Success(map[string]string{
        "token": token,
    }))
}

func main() {
    auth.SetSecret(os.Getenv("JWT_SECRET"))
    auth.SetTokenDuration(24 * time.Hour)

    mux := http.NewServeMux()
    mux.HandleFunc("/login", loginHandler)
    mux.Handle("/me", auth.RequireAuth(http.HandlerFunc(profileHandler)))

    http.ListenAndServe(":8080", mux)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    claims, _ := auth.ClaimsFromContext(r.Context())
    json.NewEncoder(w).Encode(response.Success(claims))
}
```

---

## Performance

GoKit-Lite packages are designed for high-throughput backend systems. Key packages include benchmarks you can run locally:

```bash
go test -bench=. -benchmem ./...
```

### Expected Performance Characteristics

| Package | Operation | Expected | Allocations | Benchmark File |
|---------|-----------|----------|:-----------:|----------------|
| `cache` | `Get` (hit) | ~10–50 ns/op | 0 allocs/op | `cache/cache_benchmark_test.go` |
| `cache` | `Set` | ~50–100 ns/op | 0 allocs/op | `cache/cache_benchmark_test.go` |
| `ratelimit` | `Allow` | ~10–30 ns/op | 0 allocs/op | `ratelimit/ratelimit_benchmark_test.go` |
| `config` | `Get` | ~20–50 ns/op | 0 allocs/op | `config/config_benchmark_test.go` |
| `config` | `Load` | one-time boot cost | — | `config/config_benchmark_test.go` |
| `monitor` | `RecordRequest` | lock-free atomic | 0 allocs/op | — |
| `validator` | `Validate` (5 fields) | *placeholder* | — | — |
| `auth` | `GenerateToken` | *placeholder* | — | — |

> *Placeholder* entries do not yet have dedicated benchmark tests. Contributions welcome!

See [docs/performance.md](docs/performance.md) for the full benchmarking methodology and interpretation guide.

---

## Project Goals

GoKit-Lite aims to be the **go-to utility belt** for Go backend developers who value:

- **Reusability** — write once, use across every project.
- **Readability** — clear, idiomatic Go code that's easy to understand and contribute to.
- **Maintainability** — small packages with focused responsibilities and comprehensive tests.
- **Performance** — hot paths are lock-free or use `sync.RWMutex`; zero allocations where possible.
- **Minimal dependencies** — one external dependency across all 8 packages.

---

## Project Structure

```text
GoKit-Lite/
├── auth/               # JWT authentication and middleware
│   ├── auth.go         # Manager, GenerateToken, ValidateToken
│   ├── claims.go       # Claims struct definition
│   ├── errors.go       # Sentinel error variables
│   ├── middleware.go   # RequireAuth, ClaimsFromContext
│   ├── auth_test.go
│   └── example_test.go
├── cache/              # Generic in-memory cache with TTL
│   ├── cache.go        # Cache[K,V] with Set, Get, Delete, cleanup
│   ├── cache_test.go
│   └── cache_benchmark_test.go
├── config/             # Environment variable management
│   ├── config.go       # Get (env var lookup)
│   ├── loader.go       # Load (.env file parser)
│   ├── .env.example    # Template for required settings
│   ├── config_test.go
│   ├── config_benchmark_test.go
│   └── loader_test.go
├── logger/             # Structured JSON logging
│   ├── logger.go       # Info, Error
│   └── logger_test.go
├── monitor/            # Request metrics tracking
│   ├── monitor.go      # Monitor type and package-level API
│   ├── stats.go        # Lock-free tracker implementation
│   ├── types.go        # Stats struct definition
│   ├── monitor_test.go
│   └── example_test.go
├── ratelimit/          # Token-bucket rate limiter
│   ├── ratelimit.go    # Limiter with Allow(key)
│   ├── ratelimit_test.go
│   └── ratelimit_benchmark_test.go
├── response/           # Standardized API responses
│   ├── response.go     # Success, Error
│   └── response_test.go
├── validator/          # Tag-based struct validation
│   ├── validator.go    # Validate, ValidateAll, Register
│   ├── errors.go       # ValidationError, Result
│   ├── validator_test.go
│   └── example_test.go
├── docs/               # Package documentation
│   ├── getting-started.md
│   ├── architecture.md
│   ├── performance.md
│   ├── roadmap.md
│   └── <package>.md    # Per-package guides
├── examples/           # Runnable example programs
│   ├── user-api/       # Full REST API example
│   ├── response/       # Response package demo
│   └── monitor/        # Monitoring demo
├── go.mod
├── go.sum
├── LICENSE
├── CONTRIBUTING.md
├── CHANGELOG.md
└── SECURITY.md
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Installation, prerequisites, and first steps |
| [Architecture](docs/architecture.md) | Design philosophy and package relationships |
| [Performance](docs/performance.md) | Benchmarking methodology and optimization notes |
| [Roadmap](docs/roadmap.md) | Current status and planned features |

**Package guides:** [response](docs/response.md) · [validator](docs/validator.md) · [auth](docs/auth.md) · [config](docs/config.md) · [cache](docs/cache.md) · [logger](docs/logger.md) · [monitor](docs/monitor.md) · [ratelimit](docs/ratelimit.md)

---

## Examples

| Example | Description | Location |
|---------|-------------|----------|
| User API | Full REST API with auth, validation, response, and monitoring | [`examples/user-api/`](examples/user-api/) |
| Response | Basic response envelope usage | [`examples/response/`](examples/response/) |
| Monitor | Request metrics simulation and multi-monitor demo | [`examples/monitor/`](examples/monitor/) |

Run any example:

```bash
cd examples/user-api
go run .
```

---

## Testing

```bash
# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run a specific package's tests
go test ./validator/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Roadmap

### ✅ Completed

- [x] `response` — Standardized API responses
- [x] `validator` — Tag-based struct validation (V1–V5)
- [x] `auth` — JWT generation, validation, middleware
- [x] `config` — Environment variable management with `.env` loading
- [x] `cache` — Generic in-memory cache with TTL
- [x] `logger` — Structured JSON logging
- [x] `monitor` — Lock-free request metrics tracking
- [x] `ratelimit` — Token-bucket rate limiting
- [x] Documentation suite (`docs/`)
- [x] Runnable examples (`examples/`)

### 🔜 In Progress

- [ ] CI/CD — GitHub Actions workflow for automated testing
- [ ] Repository polish — CONTRIBUTING.md, CHANGELOG.md, SECURITY.md

### 💡 Future Ideas

- [ ] Additional log levels (`DEBUG`, `WARN`, `FATAL`)
- [ ] HTTP middleware wrapper for rate limiting
- [ ] Prometheus metrics exposition in `monitor`
- [ ] Health check HTTP endpoints in `monitor`
- [ ] Contextual fields for `logger` (request ID, trace ID)
- [ ] Redis/Memcached adapter for `cache`

See the full [Roadmap](docs/roadmap.md) for details.

---

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) before submitting a pull request.

**Quick steps:**

1. Fork the repository
2. Create your branch from `develop` (`git checkout -b feature/my-feature`)
3. Write tests for your changes
4. Run `gofmt`, `go vet`, and `go test ./...`
5. Open a pull request targeting `develop`

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow, commit conventions, and coding standards.

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## Maintainers

| Name | Role | GitHub |
|------|------|--------|
| SGDevelopers | Lead Maintainer | [@sgdevelopers29-afk](https://github.com/sgdevelopers29-afk) |

**Want to contribute?** We'd love to have you! Check the [issues tab](https://github.com/sgdevelopers29-afk/GoKit-Lite/issues) for good first issues, or open one to discuss your idea. Every contribution — from typo fixes to new packages — makes GoKit-Lite better for the community.

---

<p align="center">
  Built with ❤️ in Go &nbsp;·&nbsp; <a href="https://github.com/sgdevelopers29-afk/GoKit-Lite">Star us on GitHub</a>
</p>


**sgdevelopers29-afk/GoKit-Lite** is an open-source project. Contributions are welcome! Please read the [CONTRIBUTING.md](CONTRIBUTING.md) guide before submitting a pull request.


(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)(●'◡'●)


