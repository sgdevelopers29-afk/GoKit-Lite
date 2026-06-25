// Package monitor provides lightweight, in-process request metrics for Go
// services. It tracks request counts, success/error counts, and latency using
// only the Go standard library, with no external dependencies.
//
// All operations are safe for concurrent use via sync/atomic.
//
// Basic usage:
//
//	start := time.Now()
//	monitor.RecordRequest()
//	// ... perform operation ...
//	monitor.RecordSuccess()
//	monitor.RecordLatency(time.Since(start))
//
//	stats := monitor.Stats()
//	fmt.Printf("Requests: %d, Success Rate: %.2f%%\n", stats.Requests, stats.SuccessRate)
package monitor

import "time"

// Stats holds a point-in-time snapshot of all tracked metrics.
// All values are computed at the time [Stats] or [Monitor.Stats] is called
// and represent a consistent view of the counters.
type Stats struct {
	// Requests is the total number of requests recorded via [RecordRequest].
	Requests int64 `json:"requests"`

	// Successes is the total number of successful outcomes recorded via [RecordSuccess].
	Successes int64 `json:"successes"`

	// Errors is the total number of error outcomes recorded via [RecordError].
	Errors int64 `json:"errors"`

	// SuccessRate is the ratio of Successes to Requests, expressed as a
	// percentage (0.0–100.0). It is 0.0 when Requests is zero.
	SuccessRate float64 `json:"success_rate"`

	// AverageLatency is the mean latency across all recorded samples.
	// It is zero when no latency has been recorded.
	AverageLatency time.Duration `json:"average_latency"`

	// TotalLatency is the cumulative sum of all recorded latency durations.
	TotalLatency time.Duration `json:"total_latency"`
}
