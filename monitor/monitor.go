package monitor

import "time"

// defaultTracker is the package-level tracker used by the exported
// convenience functions. It mirrors the "default instance" pattern used
// by other GoKit-Lite packages (e.g. auth.defaultManager).
var defaultTracker = &tracker{}

// ── Instance-Based API ───────────────────────────────────────────────────────

// Monitor is a named metrics collector. Use [New] to create one.
// All methods are safe for concurrent use.
type Monitor struct {
	name    string
	tracker *tracker
}

// New creates a new Monitor instance with the given name.
// The name is informational and can be used to distinguish multiple
// monitors within the same application (e.g. "http", "grpc", "db").
func New(name string) *Monitor {
	return &Monitor{
		name:    name,
		tracker: &tracker{},
	}
}

// Name returns the monitor's name.
func (m *Monitor) Name() string {
	return m.name
}

// RecordRequest increments the request counter by one.
func (m *Monitor) RecordRequest() {
	m.tracker.recordRequest()
}

// RecordSuccess increments the success counter by one.
func (m *Monitor) RecordSuccess() {
	m.tracker.recordSuccess()
}

// RecordError increments the error counter by one.
func (m *Monitor) RecordError() {
	m.tracker.recordError()
}

// RecordLatency adds the given duration to the cumulative latency total
// and increments the latency sample count. Negative durations are ignored.
func (m *Monitor) RecordLatency(d time.Duration) {
	m.tracker.recordLatency(d)
}

// Stats returns a point-in-time snapshot of this monitor's metrics.
func (m *Monitor) Stats() Stats {
	return m.tracker.snapshot()
}

// Reset zeroes all counters for this monitor.
func (m *Monitor) Reset() {
	m.tracker.reset()
}

// ── Package-Level Convenience API ────────────────────────────────────────────

// RecordRequest increments the default monitor's request counter by one.
func RecordRequest() {
	defaultTracker.recordRequest()
}

// RecordSuccess increments the default monitor's success counter by one.
func RecordSuccess() {
	defaultTracker.recordSuccess()
}

// RecordError increments the default monitor's error counter by one.
func RecordError() {
	defaultTracker.recordError()
}

// RecordLatency adds the given duration to the default monitor's cumulative
// latency total. Negative durations are ignored.
func RecordLatency(d time.Duration) {
	defaultTracker.recordLatency(d)
}

// Stats returns a point-in-time snapshot of the default monitor's metrics.
func GetStats() Stats {
	return defaultTracker.snapshot()
}

// Reset zeroes all counters on the default monitor.
func Reset() {
	defaultTracker.reset()
}
