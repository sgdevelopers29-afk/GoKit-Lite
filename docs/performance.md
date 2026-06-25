# GoKit-Lite Performance Guide

## Overview
Performance and efficiency are core tenets of the `GoKit-Lite` architecture. Because infrastructure packages (like `config`, `cache`, and `ratelimit`) are invoked thousands or millions of times per second in high-throughput backend systems, tracking their exact execution costs is critical.

This document outlines the benchmarking methodology, how to run tests, and how to interpret the results of our core infrastructure modules.

## Benchmark Methodology
We utilize Go's built-in benchmarking framework (`testing.B`). 
- Benchmarks isolate individual function calls in tight loops.
- `b.ResetTimer()` is strictly used to exclude setup allocation overhead (like writing dummy `.env` files or seeding initial maps).
- `b.RunParallel()` is used to simulate high-concurrency environments across multiple goroutines, mimicking real-world web server traffic.

## How to Run Benchmarks
To execute the complete benchmark suite and measure memory allocations, run the following command from the root of the repository:

```bash
go test -bench=. -benchmem ./...
```

To run benchmarks for a specific package, specify the path:
```bash
go test -bench=. -benchmem ./cache
```

## Interpretation of Results
When you run the benchmarks, Go provides several vital metrics:

### 1. `ns/op` (Nanoseconds per operation)
This indicates the raw execution speed of a single function call. 
- Lower is better. 
- For instance, `100 ns/op` means the operation takes 100 nanoseconds. 1,000,000 nanoseconds = 1 millisecond.

### 2. `B/op` (Bytes allocated per operation)
This tracks the amount of memory allocated on the heap during a single operation.
- Lower is better. 
- Allocating memory is expensive and triggers the Garbage Collector (GC). Zero allocations (`0 B/op`) in critical paths like `cache.Get` or `ratelimit.Allow` is the gold standard.

### 3. `allocs/op` (Allocations per operation)
This counts how many distinct heap allocations occurred during the function call.
- Lower is better.
- A function might allocate `100 B/op` across `5 allocs/op`. Reducing the number of distinct allocations drastically reduces CPU time spent in the garbage collector.

## Why Benchmarks Matter
In a backend system, an infrastructure function like `cache.Get()` might wrap every single database call. If `cache.Get` takes 500 microseconds and allocates 1KB of memory, a server handling 10,000 Requests Per Second (RPS) will suddenly allocate 10 Megabytes of garbage per second *just from the cache overhead*, stalling the server with GC pauses. Micro-optimizations at the infrastructure level compound massively at scale.

## Design Decisions
- **Config**: Relies on `os.Getenv` for single lookups resulting in near-zero overhead. `Load()` does heap allocate while splitting file strings, but is intended to run strictly once at boot.
- **Cache**: Employs `sync.RWMutex` which provides incredibly fast concurrent reads (`Get`) while safely blocking during writes (`Set`). Time complexity is O(1) amortized.
- **Rate Limiter**: Token Bucket refill logic is evaluated mathematically on-the-fly (`time.Now().Sub()`) instead of using background tickers. This design zeroes out idle memory consumption and goroutine leaks.

## Expected Performance
You should typically observe:
- **Cache Get/Set**: ~10-50 `ns/op` under normal load.
- **Rate Limit Allow**: ~10-30 `ns/op`.
- **Allocations**: `0 B/op` and `0 allocs/op` for hot paths (`cache.Get`, `ratelimit.Allow`), ensuring zero garbage collection drag during peak traffic bursts.

## Optimization Opportunities
- **Cache v4**: Switching from a global `sync.RWMutex` to a **sharded map** approach (e.g., array of 256 mutex-locked maps). This severely reduces lock contention during massive concurrent writes (`BenchmarkCacheParallelSet`).
- **Config**: Pre-loading `.env` files into a globally accessible `map[string]string` rather than pushing everything into `os.Environ()` could slightly speed up repeated lookups.
