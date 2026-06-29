// Example: monitor_example demonstrates how to use the GoKit-Lite monitor
// package to track API request metrics in a realistic HTTP handler scenario.
//
// It showcases integration with the response, validator, and auth packages.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/auth"
	"github.com/sgdevelopers29-afk/GoKit-Lite/monitor"
	"github.com/sgdevelopers29-afk/GoKit-Lite/response"
	"github.com/sgdevelopers29-afk/GoKit-Lite/validator"
)

// ── Request models ───────────────────────────────────────────────────────────

// LoginRequest represents a user login payload.
type LoginRequest struct {
	Email    string `json:"email"    required:"true" email:"true"`
	Password string `json:"password" required:"true" minLength:"8"`
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// loginHandler demonstrates a realistic API endpoint that records metrics.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	monitor.RecordRequest()

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		monitor.RecordError()
		monitor.RecordLatency(time.Since(start))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response.Error("invalid request body"))
		return
	}

	// Validate
	if err := validator.Validate(req); err != nil {
		monitor.RecordError()
		monitor.RecordLatency(time.Since(start))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response.Error(err.Error()))
		return
	}

	// Generate token
	token, err := auth.GenerateToken(auth.Claims{
		UserID: "user-123",
		Role:   "admin",
	})
	if err != nil {
		monitor.RecordError()
		monitor.RecordLatency(time.Since(start))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response.Error("token generation failed"))
		return
	}

	monitor.RecordSuccess()
	monitor.RecordLatency(time.Since(start))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.Success(map[string]string{
		"token": token,
	}))
}

// metricsHandler exposes the current monitor stats as JSON.
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	stats := monitor.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ── Simulation ───────────────────────────────────────────────────────────────

// simulateTraffic demonstrates monitoring without starting an HTTP server.
func simulateTraffic() {
	fmt.Println("=== GoKit-Lite Monitor Example ===")
	fmt.Println()

	// Create a dedicated monitor for this simulation
	apiMonitor := monitor.New("api")

	// Simulate 20 API requests with varying outcomes
	for i := 0; i < 20; i++ {
		start := time.Now()
		apiMonitor.RecordRequest()

		// Simulate processing time (1-50ms)
		time.Sleep(time.Duration(1+rand.Intn(50)) * time.Millisecond)

		// 80% success rate
		if rand.Float64() < 0.8 {
			apiMonitor.RecordSuccess()
		} else {
			apiMonitor.RecordError()
		}

		apiMonitor.RecordLatency(time.Since(start))
	}

	// Print statistics
	stats := apiMonitor.Stats()
	fmt.Printf("Monitor:         %s\n", apiMonitor.Name())
	fmt.Printf("Requests:        %d\n", stats.Requests)
	fmt.Printf("Successes:       %d\n", stats.Successes)
	fmt.Printf("Errors:          %d\n", stats.Errors)
	fmt.Printf("Success Rate:    %.2f%%\n", stats.SuccessRate)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
	fmt.Printf("Total Latency:   %v\n", stats.TotalLatency)
	fmt.Println()

	// Demonstrate reset
	fmt.Println("--- After Reset ---")
	apiMonitor.Reset()
	stats = apiMonitor.Stats()
	fmt.Printf("Requests:        %d\n", stats.Requests)
	fmt.Printf("Successes:       %d\n", stats.Successes)
	fmt.Printf("Errors:          %d\n", stats.Errors)
	fmt.Println()

	// Demonstrate multiple isolated monitors
	fmt.Println("--- Multiple Monitors ---")
	httpMon := monitor.New("http")
	dbMon := monitor.New("database")

	httpMon.RecordRequest()
	httpMon.RecordSuccess()
	httpMon.RecordLatency(15 * time.Millisecond)

	dbMon.RecordRequest()
	dbMon.RecordRequest()
	dbMon.RecordSuccess()
	dbMon.RecordError()
	dbMon.RecordLatency(5 * time.Millisecond)
	dbMon.RecordLatency(120 * time.Millisecond)

	for _, m := range []*monitor.Monitor{httpMon, dbMon} {
		s := m.Stats()
		fmt.Printf("[%s] requests=%d successes=%d errors=%d rate=%.0f%% avg_latency=%v\n",
			m.Name(), s.Requests, s.Successes, s.Errors, s.SuccessRate, s.AverageLatency)
	}
}

func main() {
	// Set up auth for the login handler demo
	auth.SetSecret("my-super-secret-key")

	// Run the simulation (no server needed)
	simulateTraffic()

	// Uncomment below to start an HTTP server with monitoring:
	//
	// http.HandleFunc("/login", loginHandler)
	// http.HandleFunc("/metrics", metricsHandler)
	// log.Println("Server starting on :8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))

	_ = loginHandler  // avoid unused warning in simulation mode
	_ = metricsHandler
	_ = log.Println
}
