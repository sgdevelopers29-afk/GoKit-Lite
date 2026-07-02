package logger_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/logger"
)

// captureOutput intercepts the logger output for testing.
func captureOutput(f func()) string {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stdout)

	f()

	return strings.TrimSpace(buf.String())
}

func TestInfo(t *testing.T) {
	msg := "test info message"
	output := captureOutput(func() {
		logger.Info(msg)
	})

	var entry logger.LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	if entry.Message != msg {
		t.Errorf("Expected message %q, got %q", msg, entry.Message)
	}
	if _, err := time.Parse(time.RFC3339, entry.Timestamp); err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got %s", entry.Timestamp)
	}
}

func TestError(t *testing.T) {
	msg := "test error message"
	output := captureOutput(func() {
		logger.Error(msg)
	})

	var entry logger.LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if entry.Level != "ERROR" {
		t.Errorf("Expected level ERROR, got %s", entry.Level)
	}
	if entry.Message != msg {
		t.Errorf("Expected message %q, got %q", msg, entry.Message)
	}
	if _, err := time.Parse(time.RFC3339, entry.Timestamp); err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got %s", entry.Timestamp)
	}
}
