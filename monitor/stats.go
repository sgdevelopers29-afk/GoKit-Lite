package monitor

import (
	"sync/atomic"
	"time"
)

// tracker is the internal, concurrency-safe metrics collector.
// Every field is accessed exclusively through atomic operations so that
// no mutex is required, giving lock-free performance under contention.
type tracker struct {
	requests      atomic.Int64
	successes     atomic.Int64
	errors        atomic.Int64
	totalLatency  atomic.Int64 // stored as nanoseconds
	latencyCount  atomic.Int64 // number of latency samples recorded
}

// recordRequest atomically increments the request counter by one.
func (t *tracker) recordRequest() {
	t.requests.Add(1)
}

// recordSuccess atomically increments the success counter by one.
func (t *tracker) recordSuccess() {
	t.successes.Add(1)
}

// recordError atomically increments the error counter by one.
func (t *tracker) recordError() {
	t.errors.Add(1)
}

// recordLatency atomically adds the given duration to the cumulative latency
// total and increments the latency sample count. Negative durations are
// silently ignored to prevent corrupted metrics.
func (t *tracker) recordLatency(d time.Duration) {
	if d < 0 {
		return
	}
	t.totalLatency.Add(int64(d))
	t.latencyCount.Add(1)
}

// snapshot returns a point-in-time [Stats] snapshot computed from the current
// counter values. Because each counter is read independently via an atomic
// load, the snapshot is not guaranteed to be perfectly consistent across all
// fields if other goroutines are mutating counters concurrently; however,
// each individual field is accurate as of its read instant.
func (t *tracker) snapshot() Stats {
	reqs := t.requests.Load()
	succ := t.successes.Load()
	errs := t.errors.Load()
	totalNs := t.totalLatency.Load()
	latCount := t.latencyCount.Load()

	var successRate float64
	if reqs > 0 {
		successRate = (float64(succ) / float64(reqs)) * 100.0
	}

	var avgLatency time.Duration
	if latCount > 0 {
		avgLatency = time.Duration(totalNs / latCount)
	}

	return Stats{
		Requests:       reqs,
		Successes:      succ,
		Errors:         errs,
		SuccessRate:    successRate,
		AverageLatency: avgLatency,
		TotalLatency:   time.Duration(totalNs),
	}
}

// reset zeroes every counter. Because each store is independent, a concurrent
// reader may observe a partially-reset state; in practice this is acceptable
// for monitoring use cases and avoids the cost of a mutex.
func (t *tracker) reset() {
	t.requests.Store(0)
	t.successes.Store(0)
	t.errors.Store(0)
	t.totalLatency.Store(0)
	t.latencyCount.Store(0)
}
