package monitor

import (
	"math"
	"sync"
	"testing"
	"time"
)

const floatEpsilon = 1e-9

// ── Package-level API tests ──────────────────────────────────────────────────

func TestRecordRequest(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordRequest()
	RecordRequest()
	RecordRequest()

	s := GetStats()
	if s.Requests != 3 {
		t.Errorf("expected 3 requests, got %d", s.Requests)
	}
}

func TestRecordSuccess(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordSuccess()
	RecordSuccess()

	s := GetStats()
	if s.Successes != 2 {
		t.Errorf("expected 2 successes, got %d", s.Successes)
	}
}

func TestRecordError(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordError()

	s := GetStats()
	if s.Errors != 1 {
		t.Errorf("expected 1 error, got %d", s.Errors)
	}
}

func TestRecordLatency(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordLatency(100 * time.Millisecond)
	RecordLatency(200 * time.Millisecond)

	s := GetStats()

	expectedTotal := 300 * time.Millisecond
	if s.TotalLatency != expectedTotal {
		t.Errorf("expected total latency %v, got %v", expectedTotal, s.TotalLatency)
	}

	expectedAvg := 150 * time.Millisecond
	if s.AverageLatency != expectedAvg {
		t.Errorf("expected average latency %v, got %v", expectedAvg, s.AverageLatency)
	}
}

func TestRecordLatencyNegativeIgnored(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordLatency(-50 * time.Millisecond)

	s := GetStats()
	if s.TotalLatency != 0 {
		t.Errorf("expected total latency 0, got %v", s.TotalLatency)
	}
}

func TestSuccessRate(t *testing.T) {
	t.Cleanup(func() { Reset() })

	// 3 requests, 2 successes = 66.67%
	RecordRequest()
	RecordRequest()
	RecordRequest()
	RecordSuccess()
	RecordSuccess()

	s := GetStats()

	expected := (2.0 / 3.0) * 100.0
	if math.Abs(s.SuccessRate-expected) > floatEpsilon {
		t.Errorf("expected success rate %.4f, got %.4f", expected, s.SuccessRate)
	}
}

func TestSuccessRateZeroRequests(t *testing.T) {
	t.Cleanup(func() { Reset() })

	s := GetStats()
	if s.SuccessRate != 0.0 {
		t.Errorf("expected success rate 0.0 with no requests, got %f", s.SuccessRate)
	}
}

func TestSuccessRateFullSuccess(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordRequest()
	RecordRequest()
	RecordSuccess()
	RecordSuccess()

	s := GetStats()
	if s.SuccessRate != 100.0 {
		t.Errorf("expected success rate 100.0, got %f", s.SuccessRate)
	}
}

func TestAverageLatencyZeroSamples(t *testing.T) {
	t.Cleanup(func() { Reset() })

	s := GetStats()
	if s.AverageLatency != 0 {
		t.Errorf("expected average latency 0 with no samples, got %v", s.AverageLatency)
	}
}

func TestReset(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordRequest()
	RecordSuccess()
	RecordError()
	RecordLatency(50 * time.Millisecond)

	Reset()

	s := GetStats()
	if s.Requests != 0 {
		t.Errorf("expected 0 requests after reset, got %d", s.Requests)
	}
	if s.Successes != 0 {
		t.Errorf("expected 0 successes after reset, got %d", s.Successes)
	}
	if s.Errors != 0 {
		t.Errorf("expected 0 errors after reset, got %d", s.Errors)
	}
	if s.TotalLatency != 0 {
		t.Errorf("expected 0 total latency after reset, got %v", s.TotalLatency)
	}
	if s.AverageLatency != 0 {
		t.Errorf("expected 0 average latency after reset, got %v", s.AverageLatency)
	}
	if s.SuccessRate != 0 {
		t.Errorf("expected 0 success rate after reset, got %f", s.SuccessRate)
	}
}

func TestResetAndReuse(t *testing.T) {
	t.Cleanup(func() { Reset() })

	RecordRequest()
	RecordSuccess()
	Reset()

	// Record new data after reset
	RecordRequest()
	RecordRequest()
	RecordError()

	s := GetStats()
	if s.Requests != 2 {
		t.Errorf("expected 2 requests after reset+reuse, got %d", s.Requests)
	}
	if s.Successes != 0 {
		t.Errorf("expected 0 successes after reset+reuse, got %d", s.Successes)
	}
	if s.Errors != 1 {
		t.Errorf("expected 1 error after reset+reuse, got %d", s.Errors)
	}
}

// ── Concurrency tests ────────────────────────────────────────────────────────

func TestConcurrentRecords(t *testing.T) {
	t.Cleanup(func() { Reset() })

	const goroutines = 100
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				RecordRequest()
				RecordSuccess()
				RecordLatency(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	s := GetStats()

	expectedOps := int64(goroutines * opsPerGoroutine)
	if s.Requests != expectedOps {
		t.Errorf("expected %d requests, got %d", expectedOps, s.Requests)
	}
	if s.Successes != expectedOps {
		t.Errorf("expected %d successes, got %d", expectedOps, s.Successes)
	}

	expectedTotalLatency := time.Duration(expectedOps) * time.Millisecond
	if s.TotalLatency != expectedTotalLatency {
		t.Errorf("expected total latency %v, got %v", expectedTotalLatency, s.TotalLatency)
	}

	if s.AverageLatency != time.Millisecond {
		t.Errorf("expected average latency %v, got %v", time.Millisecond, s.AverageLatency)
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	t.Cleanup(func() { Reset() })

	const goroutines = 50
	const opsPerGoroutine = 500

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // half success, half error

	// Success goroutines
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				RecordRequest()
				RecordSuccess()
				RecordLatency(2 * time.Millisecond)
			}
		}()
	}

	// Error goroutines
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				RecordRequest()
				RecordError()
				RecordLatency(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	s := GetStats()

	totalReqs := int64(goroutines * opsPerGoroutine * 2)
	if s.Requests != totalReqs {
		t.Errorf("expected %d requests, got %d", totalReqs, s.Requests)
	}

	expectedSuccesses := int64(goroutines * opsPerGoroutine)
	if s.Successes != expectedSuccesses {
		t.Errorf("expected %d successes, got %d", expectedSuccesses, s.Successes)
	}

	expectedErrors := int64(goroutines * opsPerGoroutine)
	if s.Errors != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, s.Errors)
	}

	if s.SuccessRate != 50.0 {
		t.Errorf("expected success rate 50.0%%, got %.2f%%", s.SuccessRate)
	}
}

func TestConcurrentResetSafety(t *testing.T) {
	// This test ensures no panics or data races occur when Reset is called
	// concurrently with recording operations.
	t.Cleanup(func() { Reset() })

	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines + 1) // recording goroutines + 1 resetter

	// Continuously record
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				RecordRequest()
				RecordSuccess()
				RecordLatency(time.Millisecond)
				_ = GetStats()
			}
		}()
	}

	// Concurrently reset
	go func() {
		defer wg.Done()
		for j := 0; j < 100; j++ {
			Reset()
		}
	}()

	wg.Wait()
	// Success = no panic and no race detector complaint
}

// ── Instance-based API tests ─────────────────────────────────────────────────

func TestNewMonitor(t *testing.T) {
	m := New("http")
	if m.Name() != "http" {
		t.Errorf("expected name 'http', got %q", m.Name())
	}
}

func TestInstanceRecordAndStats(t *testing.T) {
	m := New("test")

	m.RecordRequest()
	m.RecordRequest()
	m.RecordSuccess()
	m.RecordError()
	m.RecordLatency(100 * time.Millisecond)

	s := m.Stats()

	if s.Requests != 2 {
		t.Errorf("expected 2 requests, got %d", s.Requests)
	}
	if s.Successes != 1 {
		t.Errorf("expected 1 success, got %d", s.Successes)
	}
	if s.Errors != 1 {
		t.Errorf("expected 1 error, got %d", s.Errors)
	}
	if s.TotalLatency != 100*time.Millisecond {
		t.Errorf("expected total latency 100ms, got %v", s.TotalLatency)
	}
	if s.SuccessRate != 50.0 {
		t.Errorf("expected success rate 50.0, got %f", s.SuccessRate)
	}
}

func TestInstanceReset(t *testing.T) {
	m := New("test-reset")

	m.RecordRequest()
	m.RecordSuccess()
	m.RecordLatency(50 * time.Millisecond)
	m.Reset()

	s := m.Stats()
	if s.Requests != 0 || s.Successes != 0 || s.TotalLatency != 0 {
		t.Errorf("expected all zeros after instance reset, got %+v", s)
	}
}

func TestInstancesAreIsolated(t *testing.T) {
	a := New("service-a")
	b := New("service-b")

	a.RecordRequest()
	a.RecordRequest()
	a.RecordSuccess()

	b.RecordRequest()
	b.RecordError()

	sa := a.Stats()
	sb := b.Stats()

	if sa.Requests != 2 {
		t.Errorf("monitor-a: expected 2 requests, got %d", sa.Requests)
	}
	if sb.Requests != 1 {
		t.Errorf("monitor-b: expected 1 request, got %d", sb.Requests)
	}
	if sa.Successes != 1 {
		t.Errorf("monitor-a: expected 1 success, got %d", sa.Successes)
	}
	if sb.Errors != 1 {
		t.Errorf("monitor-b: expected 1 error, got %d", sb.Errors)
	}
}

func TestInstanceIsolatedFromDefault(t *testing.T) {
	t.Cleanup(func() { Reset() })

	m := New("isolated")
	m.RecordRequest()
	m.RecordSuccess()

	RecordRequest()
	RecordError()

	si := m.Stats()
	sd := GetStats()

	if si.Requests != 1 || si.Successes != 1 {
		t.Errorf("instance stats incorrect: %+v", si)
	}
	if sd.Requests != 1 || sd.Errors != 1 {
		t.Errorf("default stats incorrect: %+v", sd)
	}
}

// ── Benchmark ────────────────────────────────────────────────────────────────

func BenchmarkRecordRequest(b *testing.B) {
	b.Cleanup(func() { Reset() })
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RecordRequest()
	}
}

func BenchmarkRecordLatency(b *testing.B) {
	b.Cleanup(func() { Reset() })
	d := 5 * time.Millisecond
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RecordLatency(d)
	}
}

func BenchmarkGetStats(b *testing.B) {
	b.Cleanup(func() { Reset() })

	RecordRequest()
	RecordSuccess()
	RecordLatency(time.Millisecond)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetStats()
	}
}

func BenchmarkConcurrentRecord(b *testing.B) {
	b.Cleanup(func() { Reset() })
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RecordRequest()
			RecordSuccess()
			RecordLatency(time.Millisecond)
		}
	})
}
