package monitor_test

import (
	"fmt"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/monitor"
)

func ExampleRecordRequest() {
	monitor.Reset() // start clean

	monitor.RecordRequest()
	monitor.RecordRequest()
	monitor.RecordRequest()

	stats := monitor.GetStats()
	fmt.Printf("Requests: %d\n", stats.Requests)
	// Output:
	// Requests: 3
}

func ExampleRecordSuccess() {
	monitor.Reset()

	monitor.RecordRequest()
	monitor.RecordRequest()
	monitor.RecordSuccess()
	monitor.RecordSuccess()

	stats := monitor.GetStats()
	fmt.Printf("Success Rate: %.0f%%\n", stats.SuccessRate)
	// Output:
	// Success Rate: 100%
}

func ExampleRecordError() {
	monitor.Reset()

	monitor.RecordRequest()
	monitor.RecordError()

	stats := monitor.GetStats()
	fmt.Printf("Errors: %d\n", stats.Errors)
	// Output:
	// Errors: 1
}

func ExampleRecordLatency() {
	monitor.Reset()

	monitor.RecordLatency(100 * time.Millisecond)
	monitor.RecordLatency(200 * time.Millisecond)

	stats := monitor.GetStats()
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
	fmt.Printf("Total Latency: %v\n", stats.TotalLatency)
	// Output:
	// Average Latency: 150ms
	// Total Latency: 300ms
}

func ExampleGetStats() {
	monitor.Reset()

	start := time.Now()
	monitor.RecordRequest()

	// Simulate a fast operation
	time.Sleep(1 * time.Millisecond)

	monitor.RecordSuccess()
	monitor.RecordLatency(time.Since(start))

	stats := monitor.GetStats()
	fmt.Printf("Requests: %d\n", stats.Requests)
	fmt.Printf("Successes: %d\n", stats.Successes)
	fmt.Printf("Success Rate: %.0f%%\n", stats.SuccessRate)
	// Output:
	// Requests: 1
	// Successes: 1
	// Success Rate: 100%
}

func ExampleNew() {
	httpMonitor := monitor.New("http")
	grpcMonitor := monitor.New("grpc")

	// Track HTTP requests
	httpMonitor.RecordRequest()
	httpMonitor.RecordSuccess()
	httpMonitor.RecordLatency(50 * time.Millisecond)

	// Track gRPC requests separately
	grpcMonitor.RecordRequest()
	grpcMonitor.RecordError()
	grpcMonitor.RecordLatency(10 * time.Millisecond)

	httpStats := httpMonitor.Stats()
	grpcStats := grpcMonitor.Stats()

	fmt.Printf("[%s] Requests: %d, Success Rate: %.0f%%\n", httpMonitor.Name(), httpStats.Requests, httpStats.SuccessRate)
	fmt.Printf("[%s] Requests: %d, Errors: %d\n", grpcMonitor.Name(), grpcStats.Requests, grpcStats.Errors)
	// Output:
	// [http] Requests: 1, Success Rate: 100%
	// [grpc] Requests: 1, Errors: 1
}

func ExampleReset() {
	monitor.Reset()

	monitor.RecordRequest()
	monitor.RecordSuccess()

	before := monitor.GetStats()
	fmt.Printf("Before Reset - Requests: %d\n", before.Requests)

	monitor.Reset()

	after := monitor.GetStats()
	fmt.Printf("After Reset  - Requests: %d\n", after.Requests)
	// Output:
	// Before Reset - Requests: 1
	// After Reset  - Requests: 0
}
