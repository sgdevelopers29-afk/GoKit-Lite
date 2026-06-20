# Roadmap

This document tracks the current state of GoKit-Lite and outlines the planned
ecosystem of packages that contributors are working towards.

---

## Current Status

GoKit-Lite is in **active development**. The following packages are complete,
tested, and ready for use in your projects:

| Package | Version | Description | Docs |
|---|---|---|---|
| `response` | ✅ Stable | Standardised JSON envelope (`Success` / `Error`) | [response.md](response.md) |
| `validator` | ✅ Stable (V5) | Tag-based struct validation with 9 built-in rules + custom validators | [validator.md](validator.md) |
| `auth` | ✅ Stable | JWT generation, validation, and `net/http` middleware | [auth.md](auth.md) |

---

## Completed Milestones

### ✅ `response` Package
- `Response` struct with `Success bool`, `Message string`, `Data any` fields
- `Success(data any) Response` — uniform success envelope
- `Error(message string) Response` — uniform error envelope
- `"data"` field omitted from JSON when `nil` (`omitempty`)

### ✅ `validator` Package (V1 → V5 progression)
- **V1:** `required` rule
- **V2:** `min`, `max` rules for numeric fields
- **V3:** `email`, `minLength`, `maxLength` rules
- **V4:** `regex`, `oneOf`, `eqField` rules
- **V5 (current):**
  - `ValidateAll` — aggregate all field errors in one pass
  - `Result` type with `Valid bool` and `[]ValidationError`
  - `Register` — custom field-level validators
  - `RegisterStructValidator` — cross-field struct-level validators
  - Recursive validation of nested structs, slices, and maps
  - Regex pattern caching for performance

### ✅ `auth` Package
- `Claims` struct with `UserID`, `Email`, `Role`, and embedded `jwt.RegisteredClaims`
- `Manager` type for scoped token management (separate secrets / durations)
- `GenerateToken` — HMAC-SHA256 signed JWT with automatic `iat` / `exp`
- `ValidateToken` — signature verification + expiration check
- `RequireAuth` — standard `net/http` middleware
- `ClaimsFromContext` — safe claims extraction from request context
- Sentinel error variables for precise error handling
- Package-level wrappers for single-manager convenience

---

## Planned Modules

The following packages are planned by the broader GoKit-Lite contributor team.
They are listed here to give the community visibility into the project's
direction. **Implementation details for these packages are not yet finalised.**

> ℹ️ These modules are owned by other contributors. Do not assume API
> stability or implementation details at this stage.

---

### 🔜 `cache`

An in-memory (and optionally distributed) caching layer designed for common
backend patterns:

- Simple `Get` / `Set` / `Delete` operations
- TTL (time-to-live) support for automatic expiry
- Optional Redis / Memcached backend adapters
- Thread-safe for concurrent access

**Planned use case:**
```go
// Cache a user object for 5 minutes
cache.Set("user:usr_001", userObject, 5*time.Minute)
user, found := cache.Get("user:usr_001")
```

---

### 🔜 `logger`

A structured, zero-allocation logger for production observability:

- JSON output for log aggregation systems (Datadog, ELK, Cloud Logging)
- Log levels: `DEBUG`, `INFO`, `WARN`, `ERROR`
- Contextual fields (request ID, user ID, trace ID)
- Drop-in friendly interface

**Planned use case:**
```go
logger.Info("user registered", "user_id", "usr_001", "email", "alice@example.com")
logger.Error("database query failed", "err", err, "query", "SELECT ...")
```

---

### 🔜 `monitor`

Lightweight application health-check and metrics utilities:

- `/healthz` and `/readyz` HTTP endpoints
- Basic runtime metrics (goroutine count, memory usage, GC stats)
- Optional Prometheus metrics exposition
- Uptime tracking

**Planned use case:**
```go
monitor.RegisterHealth("database", func() error {
    return db.Ping()
})
http.Handle("/healthz", monitor.HealthHandler())
```

---

### 🔜 `ratelimit`

Token-bucket and sliding-window rate limiting middleware:

- Per-IP and per-user-ID rate limiting
- Configurable burst and sustained rate
- HTTP middleware compatible with `net/http`
- Pluggable storage backends (in-memory, Redis)

**Planned use case:**
```go
limiter := ratelimit.New(ratelimit.Config{
    Rate:  100,           // 100 requests
    Per:   time.Minute,   // per minute
    Burst: 20,            // allow short bursts
})
mux.Handle("/api/", ratelimit.Middleware(limiter)(apiRouter))
```

---

## Contributing

We welcome contributions! Whether you're fixing a bug, improving documentation,
or implementing a planned module, here's how to get started:

### Good First Issues

- Improving error messages in `validator`
- Adding more runnable examples to `example_test.go` files
- Writing benchmarks for the `validator` hot path
- Improving README clarity for beginners

### Starting a New Module

1. Check the [issues tab](https://github.com/sgdevelopers29-afk/GoKit-Lite/issues)
   to see if a module is already being worked on.
2. Open an issue to discuss your design before writing code.
3. Create a branch from `develop`: `git checkout -b feature/<package-name>`.
4. Follow the conventions established by the existing packages:
   - One directory per package
   - `package <name>` at the top of every file
   - Exported functions documented with standard Go doc comments
   - A `<package>_test.go` with unit tests
   - An `example_test.go` with runnable `Example*` functions
   - A `docs/<package>.md` documentation file
5. Open a PR targeting `develop`.

### Contribution Guidelines

- **Branching:** Branch from `develop` → `feature/...` or `fix/...`
- **Tests:** Run `go test ./...` before submitting. New code requires tests.
- **Code style:** `gofmt` + `go vet`. Keep changes focused.
- **Commits:** Use clear, imperative commit messages.
- **License:** All contributions are licensed under MIT.

---

## Version History

| Version / Milestone | Notable Changes |
|---|---|
| Initial release | `response`, `validator` (V1: `required`) |
| Validator V2 | `min`, `max` numeric rules |
| Validator V3 | `email`, `minLength`, `maxLength` |
| Validator V4 | `regex`, `oneOf`, `eqField` |
| Validator V5 | `ValidateAll`, `Result`, custom validators, struct validators |
| Auth release | `GenerateToken`, `ValidateToken`, `RequireAuth` middleware |
| Examples | `examples/user-api` — complete runnable example |
| Docs | Full documentation suite (`docs/`) |
