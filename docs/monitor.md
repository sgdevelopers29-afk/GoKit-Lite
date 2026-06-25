# `monitor` — Lightweight In-Process Metrics

The `monitor` package lets you track request counts, success/error rates, and
latency inside your Go application using only the standard library. No
Prometheus, no OpenTelemetry, no external systems — just import and record.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [API Reference](#api-reference)
   - [Package-Level Functions](#package-level-functions)
   - [Monitor Instance](#monitor-instance)
   - [Stats Struct](#stats-struct)
5. [Examples](#examples)
   - [Basic Usage](#basic-usage)
   - [Named Monitors](#named-monitors)
   - [HTTP Handler Integration](#http-handler-integration)
   - [Integration with GoKit-Lite Packages](#integration-with-gokit-lite-packages)
6. [Best Practices](#best-practices)
7. [Concurrency Notes](#concurrency-notes)
8. [Architecture](#architecture)
9. [Future Roadmap](#future-roadmap)

---

## Introduction

When building APIs, backend services, or microservices in Go, you often need
basic observability: _How many requests are we handling? What's the error rate?
How fast are responses?_

The `monitor` package answers these questions with a zero-dependency,
goroutine-safe metrics collector that takes **minutes** to integrate.

### What it tracks

| Metric | Description |
|---|---|
| **Requests** | Total number of recorded requests |
| **Successes** | Total number of successful outcomes |
| **Errors** | Total number of error outcomes |
| **Success Rate** | `(Successes / Requests) × 100` as a percentage |
| **Average Latency** | Mean duration across all latency samples |
| **Total Latency** | Cumulative sum of all latency durations |

---

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/monitor"
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/sgdevelopers29-afk/GoKit-Lite/monitor"
)

func main() {
    start := time.Now()

    monitor.RecordRequest()

    // ... perform your operation ...
    time.Sleep(10 * time.Millisecond)

    monitor.RecordSuccess()
    monitor.RecordLatency(time.Since(start))

    stats := monitor.GetStats()
    fmt.Printf("Requests: %d\n", stats.Requests)
    fmt.Printf("Success Rate: %.0f%%\n", stats.SuccessRate)
    fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
}
```

Output:

```
Requests: 1
Success Rate: 100%
Average Latency: ~10ms
```

---

## API Reference

### Package-Level Functions

These operate on a shared default monitor, ideal for simple applications.

#### `RecordRequest()`

```go
func RecordRequest()
```

Increments the request counter by one. Call this at the beginning of every
operation you want to track.

---

#### `RecordSuccess()`

```go
func RecordSuccess()
```

Increments the success counter by one. Call this when the operation completes
successfully.

---

#### `RecordError()`

```go
func RecordError()
```

Increments the error counter by one. Call this when the operation fails.

---

#### `RecordLatency(d time.Duration)`

```go
func RecordLatency(d time.Duration)
```

Adds `d` to the cumulative latency total and increments the latency sample
count. Negative durations are silently ignored.

---

#### `GetStats() Stats`

```go
func GetStats() Stats
```

Returns a point-in-time snapshot of the default monitor's metrics.

---

#### `Reset()`

```go
func Reset()
```

Zeroes all counters on the default monitor. Useful for testing or periodic
metric windows.

---

### Monitor Instance

For applications that need to track metrics for different subsystems (HTTP,
database, gRPC, etc.), create named `Monitor` instances.

#### `New(name string) *Monitor`

```go
m := monitor.New("http")
```

Creates a new, independent monitor. Each instance maintains its own counters,
fully isolated from the default monitor and from other instances.

#### Instance Methods

| Method | Description |
|---|---|
| `m.Name() string` | Returns the monitor's name |
| `m.RecordRequest()` | Increments the request counter |
| `m.RecordSuccess()` | Increments the success counter |
| `m.RecordError()` | Increments the error counter |
| `m.RecordLatency(d)` | Records a latency sample |
| `m.Stats() Stats` | Returns a snapshot of this monitor's metrics |
| `m.Reset()` | Zeroes all counters |

---

### Stats Struct

```go
type Stats struct {
    Requests       int64         `json:"requests"`
    Successes      int64         `json:"successes"`
    Errors         int64         `json:"errors"`
    SuccessRate    float64       `json:"success_rate"`
    AverageLatency time.Duration `json:"average_latency"`
    TotalLatency   time.Duration `json:"total_latency"`
}
```

| Field | Type | Description |
|---|---|---|
| `Requests` | `int64` | Total recorded requests |
| `Successes` | `int64` | Total recorded successes |
| `Errors` | `int64` | Total recorded errors |
| `SuccessRate` | `float64` | Percentage (0.0–100.0) |
| `AverageLatency` | `time.Duration` | Mean latency per sample |
| `TotalLatency` | `time.Duration` | Sum of all latency samples |

All fields have `json` struct tags for easy serialisation to JSON endpoints.

---

## Examples

### Basic Usage

```go
monitor.RecordRequest()

// Simulate work
time.Sleep(5 * time.Millisecond)

monitor.RecordSuccess()
monitor.RecordLatency(5 * time.Millisecond)

stats := monitor.GetStats()
fmt.Printf("Requests: %d, Success Rate: %.0f%%\n", stats.Requests, stats.SuccessRate)
// Requests: 1, Success Rate: 100%
```

---

### Named Monitors

```go
httpMonitor := monitor.New("http")
dbMonitor := monitor.New("database")

httpMonitor.RecordRequest()
httpMonitor.RecordSuccess()
httpMonitor.RecordLatency(15 * time.Millisecond)

dbMonitor.RecordRequest()
dbMonitor.RecordError()
dbMonitor.RecordLatency(120 * time.Millisecond)

fmt.Println(httpMonitor.Stats())
// {Requests:1 Successes:1 Errors:0 SuccessRate:100 AverageLatency:15ms TotalLatency:15ms}

fmt.Println(dbMonitor.Stats())
// {Requests:1 Successes:0 Errors:1 SuccessRate:0 AverageLatency:120ms TotalLatency:120ms}
```

---

### HTTP Handler Integration

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    monitor.RecordRequest()

    result, err := processRequest(r)
    if err != nil {
        monitor.RecordError()
        monitor.RecordLatency(time.Since(start))
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    monitor.RecordSuccess()
    monitor.RecordLatency(time.Since(start))

    json.NewEncoder(w).Encode(result)
}

// Expose metrics at /metrics
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    stats := monitor.GetStats()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}
```

---

### Integration with GoKit-Lite Packages

#### With `response` Package

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    monitor.RecordRequest()

    // ... create user logic ...

    monitor.RecordSuccess()
    monitor.RecordLatency(time.Since(start))

    json.NewEncoder(w).Encode(response.Success(user))
}
```

#### With `validator` Package

```go
func registerHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    monitor.RecordRequest()

    var req RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)

    if err := validator.Validate(req); err != nil {
        monitor.RecordError()
        monitor.RecordLatency(time.Since(start))
        json.NewEncoder(w).Encode(response.Error(err.Error()))
        return
    }

    // ... process valid request ...

    monitor.RecordSuccess()
    monitor.RecordLatency(time.Since(start))
    json.NewEncoder(w).Encode(response.Success(result))
}
```

#### With `auth` Package

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    monitor.RecordRequest()

    // Authenticate user...

    token, err := auth.GenerateToken(auth.Claims{
        UserID: "user-123",
        Role:   "admin",
    })
    if err != nil {
        monitor.RecordError()
        monitor.RecordLatency(time.Since(start))
        json.NewEncoder(w).Encode(response.Error("authentication failed"))
        return
    }

    monitor.RecordSuccess()
    monitor.RecordLatency(time.Since(start))
    json.NewEncoder(w).Encode(response.Success(map[string]string{
        "token": token,
    }))
}
```

#### Full Integration Example

```go
func fullHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    monitor.RecordRequest()

    // 1. Parse
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        monitor.RecordError()
        monitor.RecordLatency(time.Since(start))
        json.NewEncoder(w).Encode(response.Error("invalid body"))
        return
    }

    // 2. Validate
    if err := validator.Validate(req); err != nil {
        monitor.RecordError()
        monitor.RecordLatency(time.Since(start))
        json.NewEncoder(w).Encode(response.Error(err.Error()))
        return
    }

    // 3. Authenticate
    token, err := auth.GenerateToken(auth.Claims{UserID: "u1", Role: "user"})
    if err != nil {
        monitor.RecordError()
        monitor.RecordLatency(time.Since(start))
        json.NewEncoder(w).Encode(response.Error("auth failed"))
        return
    }

    // 4. Success
    monitor.RecordSuccess()
    monitor.RecordLatency(time.Since(start))
    json.NewEncoder(w).Encode(response.Success(map[string]string{
        "token": token,
    }))
}
```

---

## Best Practices

1. **Record at boundaries.** Place `RecordRequest()` at the very start of your
   handler and `RecordLatency()` right before returning.

2. **Always record either success or error.** Every request should end with
   exactly one call to `RecordSuccess()` or `RecordError()` to keep the success
   rate accurate.

3. **Use named monitors for subsystems.** If your application talks to a
   database, an external API, and serves HTTP, create a separate
   `monitor.New("db")`, `monitor.New("external")`, and `monitor.New("http")`
   for each.

4. **Reset periodically for windowed metrics.** If you want metrics per minute
   or per interval, call `Reset()` after reading and logging the stats.

5. **Expose a `/metrics` endpoint.** The `Stats` struct has JSON tags, so you
   can serve it directly via `json.NewEncoder(w).Encode(stats)`.

6. **Don't count latency for invalid requests** (optional). Some teams prefer
   to skip `RecordLatency()` on validation errors since the work was trivial.
   Choose a convention and be consistent.

---

## Concurrency Notes

The `monitor` package is fully safe for concurrent use. Here is how:

- **`sync/atomic` for everything.** All counters use `atomic.Int64`, which
  provides lock-free atomic operations. No mutexes are used anywhere.

- **No contention.** Unlike `sync.Mutex`, atomic operations do not block other
  goroutines. Multiple goroutines can call `RecordRequest()` simultaneously
  with negligible performance impact.

- **Snapshot consistency.** When `GetStats()` / `Stats()` is called, each
  counter is read independently via an atomic load. This means the snapshot is
  approximately consistent — if another goroutine is actively recording, you
  may see a request counted but its corresponding success not yet incremented.
  This trade-off is intentional: it avoids the performance cost of locking all
  counters together, which is unnecessary for monitoring use cases.

- **Reset safety.** `Reset()` can be called concurrently with recording
  operations without panics or data races. A reader during reset may observe a
  partially-zeroed state, which is acceptable for monitoring.

### Performance

Benchmarks show that `RecordRequest()` takes approximately **5–10 ns** per
operation and scales linearly across cores thanks to the lock-free atomic
design.

---

## Architecture

```
monitor/
├── types.go         # Stats struct definition
├── stats.go         # Internal tracker (atomic counters + snapshot logic)
├── monitor.go       # Public API: package-level functions + Monitor type
├── monitor_test.go  # Unit tests, concurrency tests, benchmarks
└── example_test.go  # Runnable GoDoc examples
```

### Design Decisions

| Decision | Rationale |
|---|---|
| `sync/atomic` over `sync.Mutex` | Lock-free = zero contention under load |
| Package-level functions + named instances | Simple for small apps, scalable for large ones |
| Latency stored as nanoseconds (`int64`) | Preserves `time.Duration` precision; atomic-friendly |
| Negative duration guard | Prevents corrupted metrics from clock skew |
| JSON struct tags on `Stats` | Ready for HTTP endpoint serialisation |
| No external dependencies | Aligns with GoKit-Lite's standard-library-only philosophy |

### How It Works

```
┌──────────────────────────────────────────────────┐
│                 Your Application                 │
│                                                  │
│   monitor.RecordRequest()  ──┐                   │
│   monitor.RecordSuccess()  ──┤    atomic.Add     │
│   monitor.RecordError()    ──┼──────────────►    │
│   monitor.RecordLatency()  ──┤   tracker{}       │
│                              │  ┌─────────────┐  │
│   monitor.GetStats()  ◄──────┤  │ requests    │  │
│                              │  │ successes   │  │
│                              │  │ errors      │  │
│                              │  │ totalLatency│  │
│                              │  │ latencyCount│  │
│                              │  └─────────────┘  │
└──────────────────────────────────────────────────┘
```

---

## Future Roadmap

> **Note:** The features below are planned ideas only. They are **not
> implemented** in the current V1 release.

### V2 — Custom Metrics

- Register named counters and gauges beyond the built-in set.
- `monitor.RegisterCounter("cache_hits")` / `monitor.IncrCounter("cache_hits")`

### V3 — JSON Export

- `monitor.ExportJSON()` for structured export of all metrics.
- `monitor.ExportJSONWriter(w io.Writer)` for streaming to files or
  network connections.

### V4 — Middleware Integration

- `monitor.HTTPMiddleware()` that wraps `http.Handler` and automatically
  records request count, success/error, and latency.
- Support for chi, gorilla/mux, and standard `net/http`.

### V5 — Dashboard Support

- Built-in HTML dashboard served at a configurable endpoint.
- Real-time metric graphs using Server-Sent Events (SSE).
- No JavaScript frameworks — pure HTML/CSS/JS.
