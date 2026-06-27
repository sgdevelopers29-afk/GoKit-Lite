# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-06-27

### Added

#### `response` Package
- `Response` struct with standardized JSON envelope (`Success`, `Message`, `Data`).
- `Success(data any) Response` — creates a uniform success response.
- `Error(message string) Response` — creates a uniform error response.
- `Data` field uses `omitempty` to omit from JSON when nil.
- Unit tests with full coverage.

#### `validator` Package (V1–V5)
- **V1:** `required` tag — ensures fields are not zero-valued.
- **V2:** `min`, `max` tags — numeric range validation for int, uint, and float types.
- **V3:** `email`, `minLength`, `maxLength` tags — string format and length rules.
- **V4:** `regex`, `oneOf`, `eqField` tags — pattern matching, enum constraints, and cross-field equality.
- **V5:** Full validation framework:
  - `Validate(data any) error` — fail-fast single-error validation.
  - `ValidateAll(data any) *Result` — aggregate all field errors in one pass.
  - `Register(name, fn)` — custom field-level validators with tag activation.
  - `RegisterStructValidator(type, fn)` — cross-field struct-level validators.
  - `Unregister` / `UnregisterStructValidator` — teardown support.
  - Recursive validation of nested structs, slices, arrays, and maps.
  - `ValidationError` type with `Field`, `Rule`, and `Message` fields.
  - `Result` type with `Valid`, `Errors`, `Error()`, and `First()`.
  - Regex pattern caching via `sync.Map` for performance.
  - Maximum recursion depth guard (32 levels).
- Comprehensive unit tests (52 KB) and example tests.

#### `auth` Package
- `Claims` struct with `UserID`, `Email`, `Role`, and embedded `jwt.RegisteredClaims`.
- `Manager` type for scoped token management with configurable secrets and durations.
- `NewManager(secret, duration)` — constructor with safe defaults.
- `GenerateToken(Claims) (string, error)` — HMAC-SHA256 signed JWT with automatic `iat`/`exp`.
- `ValidateToken(tokenString) (*Claims, error)` — signature verification and expiration check.
- `RequireAuth(next http.Handler) http.Handler` — standard `net/http` middleware.
- `ClaimsFromContext(ctx) (*Claims, bool)` — safe claims extraction from request context.
- `SetSecret`, `SetTokenDuration` — package-level configuration helpers.
- Sentinel error variables: `ErrInvalidToken`, `ErrExpiredToken`, `ErrInvalidSignature`, `ErrMissingSecret`, `ErrMissingAuthHeader`, `ErrInvalidAuthHeader`, `ErrInvalidClaims`.
- Package-level wrappers using a default `Manager` instance.
- Unit tests and example tests.

#### `config` Package
- `Get(key string) string` — centralized environment variable lookup.
- `Load(path string) error` — `.env` file parser with `KEY=VALUE` format.
- Supports comments (`#`) and blank lines.
- Whitespace trimming around keys and values.
- Line-number-aware error reporting for invalid formats.
- Unit tests and benchmarks.

#### `cache` Package
- `Cache[K comparable, V any]` — generic, thread-safe in-memory cache.
- `New[K, V]()` — constructor returning an initialized cache instance.
- `Set(key, value)` — store with no expiration.
- `SetWithTTL(key, value, ttl)` — store with time-to-live.
- `Get(key) (V, bool)` — retrieve with lazy expiration eviction.
- `Has(key) bool` — existence check.
- `Delete(key)` — remove a single entry.
- `Clear()` — remove all entries.
- `Size() int` — current entry count.
- `StartCleanup(interval)` — background goroutine for expired entry eviction.
- `StopCleanup()` — graceful cleanup worker shutdown.
- `DeleteExpired()` — manual sweep of expired entries.
- Double-check locking pattern in `Get` for safe concurrent access.
- Unit tests and benchmarks.

#### `logger` Package
- `LogEntry` struct with `Level`, `Message`, and `Timestamp` fields.
- `Info(message string)` — structured JSON log at INFO level.
- `Error(message string)` — structured JSON log at ERROR level.
- RFC 3339 UTC timestamps.
- JSON marshaling with graceful fallback.
- Unit tests.

#### `monitor` Package
- `Monitor` type — named metrics collector with instance-level isolation.
- `New(name string) *Monitor` — constructor for named monitors.
- `RecordRequest()`, `RecordSuccess()`, `RecordError()` — event counters.
- `RecordLatency(d time.Duration)` — latency tracking (negative durations ignored).
- `Stats() Stats` — point-in-time metrics snapshot.
- `Reset()` — zero all counters.
- `Stats` struct with `Requests`, `Successes`, `Errors`, `SuccessRate`, `AverageLatency`, `TotalLatency`.
- Lock-free implementation using `sync/atomic`.
- Package-level convenience API via default tracker.
- Unit tests, example tests, and comprehensive edge-case coverage.

#### `ratelimit` Package
- `Limiter` — in-memory token-bucket rate limiter.
- `New(rate, burst int) *Limiter` — constructor with tokens-per-second and burst capacity.
- `Allow(key string) bool` — per-key rate limit check with lazy bucket initialization.
- Mathematical token refill based on elapsed time (no background goroutine).
- Thread-safe via `sync.RWMutex`.
- Unit tests and benchmarks.

#### Documentation
- `docs/getting-started.md` — installation and first steps guide.
- `docs/architecture.md` — design philosophy and package relationships.
- `docs/performance.md` — benchmarks and optimization notes.
- `docs/roadmap.md` — project status and planned features.
- Per-package documentation: `response.md`, `validator.md`, `auth.md`, `config.md`, `cache.md`, `logger.md`, `monitor.md`, `ratelimit.md`.

#### Examples
- `examples/user-api/` — complete REST API demonstrating auth, validation, response, and monitoring.
- `examples/response/` — basic response envelope usage.
- `examples/monitor/` — metrics simulation and multi-monitor demo.

[1.0.0]: https://github.com/sgdevelopers29-afk/GoKit-Lite/releases/tag/v1.0.0
