package config

import (
	"os"
	"testing"
)

// TestGet verifies that Get correctly retrieves environment variable values.
//
// Test cases:
//   - Existing environment variable should return its value.
//   - Missing environment variable should return an empty string.
func TestGet(t *testing.T) {
	testKey := "TEST_ENV_VAR"
	testValue := "test_value"

	tests := []struct {
		name     string
		key      string
		setup    func(t *testing.T)
		expected string
	}{
		{
			name: "Existing environment variable",
			key:  testKey,
			setup: func(t *testing.T) {
				if err := os.Setenv(testKey, testValue); err != nil {
					t.Fatalf("failed to set env: %v", err)
				}
			},
			expected: testValue,
		},
		{
			name:     "Missing environment variable",
			key:      "NON_EXISTENT_VAR",
			setup:    func(t *testing.T) {},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			// Ensure test environment variables are cleaned up
			// after each test case execution.
			t.Cleanup(func() {
				os.Unsetenv(testKey)
			})

			got := Get(tt.key)

			if got != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}
