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

Most Go backend projects end up re-implementing the same building blocks — standardized API responses, input validation, JWT authentication, caching, rate limiting, and monitoring. GoKit-Lite packages these recurring patterns into a cohesive, zero-bloat toolkit that stays out of your way.

**Design principles:**

- **Zero framework lock-in** — every package works with `net/http` and the standard library.
- **Minimal dependencies** — only [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt) for the `auth` package; everything else is stdlib-only.
- **Pick what you need** — import a single package without pulling in the entire toolkit.
- **Production-ready defaults** — thread-safe, well-tested, and benchmarked.

---

## Features

| Package | Description | Docs |
|---------|-------------|------|
| [`response`](response/) | Standardized JSON API envelope (`Success` / `Error`) | [response.md](docs/response.md) |
| [`validator`](validator/) | Tag-based struct validation with 9 built-in rules, custom validators, and recursive nested support | [validator.md](docs/validator.md) |
| [`auth`](auth/) | JWT generation, validation, and `net/http` middleware (HMAC-SHA256) | [auth.md](docs/auth.md) |
| [`config`](config/) | Environment variable management with `.env` file loading | [config.md](docs/config.md) |
| [`cache`](cache/) | Generic, thread-safe in-memory cache with TTL and background cleanup | [cache.md](docs/cache.md) |
| [`logger`](logger/) | Simple, structured JSON logging (`INFO` / `ERROR`) | [logger.md](docs/logger.md) |
| [`monitor`](monitor/) | Lock-free request metrics — counts, success rate, and latency tracking | [monitor.md](docs/monitor.md) |
| [`ratelimit`](ratelimit/) | Token-bucket rate limiter with per-key tracking | [ratelimit.md](docs/ratelimit.md) |

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

## Project Structure

```text
GoKit-Lite/
├── auth/               # JWT authentication and middleware
│   ├── auth.go         # Manager, GenerateToken, ValidateToken
│   ├── claims.go       # Claims struct definition
│   ├── errors.go       # Sentinel error variables
│   └── middleware.go   # RequireAuth, ClaimsFromContext
├── cache/              # Generic in-memory cache with TTL
│   └── cache.go        # Cache[K,V] with Set, Get, Delete, cleanup
├── config/             # Environment variable management
│   ├── config.go       # Get (env var lookup)
│   └── loader.go       # Load (.env file parser)
├── logger/             # Structured JSON logging
│   └── logger.go       # Info, Error
├── monitor/            # Request metrics tracking
│   ├── monitor.go      # Monitor type and package-level API
│   ├── stats.go        # Lock-free tracker implementation
│   └── types.go        # Stats struct definition
├── ratelimit/          # Token-bucket rate limiter
│   └── ratelimit.go    # Limiter with Allow(key)
├── response/           # Standardized API responses
│   └── response.go     # Success, Error
├── validator/          # Tag-based struct validation
│   ├── validator.go    # Validate, ValidateAll, Register
│   └── errors.go       # ValidationError, Result
├── docs/               # Package documentation
│   ├── getting-started.md
│   ├── architecture.md
│   ├── performance.md
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
├── CODE_OF_CONDUCT.md
└── SECURITY.md
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and first steps |
| [Architecture](docs/architecture.md) | Design philosophy and package relationships |
| [Performance](docs/performance.md) | Benchmarks and optimization notes |
| [Roadmap](docs/roadmap.md) | Current status and planned features |

**Package guides:** [response](docs/response.md) · [validator](docs/validator.md) · [auth](docs/auth.md) · [config](docs/config.md) · [cache](docs/cache.md) · [logger](docs/logger.md) · [monitor](docs/monitor.md) · [ratelimit](docs/ratelimit.md)

---

## Examples

| Example | Description | Location |
|---------|-------------|----------|
| User API | Full REST API with auth, validation, and monitoring | [`examples/user-api/`](examples/user-api/) |
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

- [x] `response` — Standardized API responses
- [x] `validator` — Tag-based struct validation (V1–V5)
- [x] `auth` — JWT generation, validation, middleware
- [x] `config` — Environment variable management
- [x] `cache` — Generic in-memory cache with TTL
- [x] `logger` — Structured JSON logging
- [x] `monitor` — Request metrics tracking
- [x] `ratelimit` — Token-bucket rate limiting
- [ ] CI/CD — GitHub Actions workflow
- [ ] Additional log levels (`DEBUG`, `WARN`)
- [ ] HTTP middleware for rate limiting
- [ ] Prometheus metrics exposition

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

---

<p align="center">
  Built with ❤️ in Go &nbsp;·&nbsp; <a href="https://github.com/sgdevelopers29-afk/GoKit-Lite">Star us on GitHub</a>
</p>
