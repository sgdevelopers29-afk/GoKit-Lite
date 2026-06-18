package logger_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/logger"
)

// resetLogger restores default package-level settings after each test.
func resetLogger() {
	logger.SetLevel(logger.LevelInfo)
	logger.SetOutput(os.Stdout)
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected logger.Level
	}{
		{"TRACE", logger.LevelTrace},
		{"debug", logger.LevelDebug},
		{"  Info  ", logger.LevelInfo},
		{"WARN", logger.LevelWarn},
		{"Error", logger.LevelError},
		{"UNKNOWN", logger.LevelInfo}, // default fallback
		{"", logger.LevelInfo},        // empty fallback
	}

	for _, tc := range tests {
		result := logger.ParseLevel(tc.input)
		if result != tc.expected {
			t.Errorf("ParseLevel(%q) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		lvl      logger.Level
		expected string
	}{
		{logger.LevelTrace, "TRACE"},
		{logger.LevelDebug, "DEBUG"},
		{logger.LevelInfo, "INFO"},
		{logger.LevelWarn, "WARN"},
		{logger.LevelError, "ERROR"},
		{logger.Level(99), "UNKNOWN"},
	}

	for _, tc := range tests {
		result := tc.lvl.String()
		if result != tc.expected {
			t.Errorf("Level(%d).String() = %q; want %q", tc.lvl, result, tc.expected)
		}
	}
}

func TestJSONOutputFormatAndLevels(t *testing.T) {
	defer resetLogger()

	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetLevel(logger.LevelTrace)

	tests := []struct {
		logFunc func(string)
		level   string
		message string
	}{
		{logger.Trace, "TRACE", "trace msg"},
		{logger.Debug, "DEBUG", "debug msg"},
		{logger.Info, "INFO", "info msg"},
		{logger.Warn, "WARN", "warn msg"},
		{logger.Error, "ERROR", "error msg"},
	}

	for _, tc := range tests {
		buf.Reset()
		tc.logFunc(tc.message)

		var entry logger.LogEntry
		err := json.Unmarshal(buf.Bytes(), &entry)
		if err != nil {
			t.Fatalf("Failed to parse JSON output: %v; Raw: %q", err, buf.String())
		}

		if entry.Level != tc.level {
			t.Errorf("Expected level %q, got %q", tc.level, entry.Level)
		}
		if entry.Message != tc.message {
			t.Errorf("Expected message %q, got %q", tc.message, entry.Message)
		}
		if _, err := time.Parse(time.RFC3339, entry.Timestamp); err != nil {
			t.Errorf("Expected valid RFC3339 timestamp, got %q", entry.Timestamp)
		}
	}
}

func TestLevelFiltering(t *testing.T) {
	defer resetLogger()

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Set level to WARN. INFO, DEBUG, TRACE should be filtered out. WARN and ERROR should pass.
	logger.SetLevel(logger.LevelWarn)

	logger.Trace("should not print")
	logger.Debug("should not print")
	logger.Info("should not print")

	if buf.Len() > 0 {
		t.Errorf("Expected filtered messages to write nothing, but got: %q", buf.String())
	}

	logger.Warn("should print warn")
	logger.Error("should print error")

	output := buf.String()
	if !strings.Contains(output, `"level":"WARN"`) || !strings.Contains(output, `"level":"ERROR"`) {
		t.Errorf("Expected WARN and ERROR logs to be present, got: %q", output)
	}
}

func TestSetOutputFile(t *testing.T) {
	defer resetLogger()

	tmpFile, err := os.CreateTemp("", "gokit_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	err = logger.SetOutputFile(tmpFilePath)
	if err != nil {
		t.Fatalf("SetOutputFile failed: %v", err)
	}

	testMsg := "hello file logging"
	logger.Info(testMsg)

	// Read content
	content, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	var entry logger.LogEntry
	err = json.Unmarshal(content, &entry)
	if err != nil {
		t.Fatalf("File content not valid JSON: %v; Content: %q", err, string(content))
	}

	if entry.Message != testMsg {
		t.Errorf("Expected message %q in file, got %q", testMsg, entry.Message)
	}
}

func TestSetOutputFileError(t *testing.T) {
	defer resetLogger()
	// Pass an invalid path that cannot be created
	err := logger.SetOutputFile("/nonexistent_dir/no_permission.log")
	if err == nil {
		t.Error("Expected error when setting output to a nonexistent directory, but got nil")
	}
}

func TestConcurrentLoggingSafety(t *testing.T) {
	defer resetLogger()

	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetLevel(logger.LevelDebug)

	var wg sync.WaitGroup
	workers := 50
	logsPerWorker := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < logsPerWorker; j++ {
				// Mix different level calls
				logger.Debug(strings.Repeat("D", 10))
				logger.Info(strings.Repeat("I", 10))
				logger.Warn(strings.Repeat("W", 10))
			}
		}(i)
	}

	wg.Wait()

	// Parse JSON lines from the output buffer
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	expectedCount := workers * logsPerWorker * 3

	if len(lines) != expectedCount {
		t.Errorf("Expected %d log lines, got %d", expectedCount, len(lines))
	}

	// Verify each line is valid JSON (no partial or corrupted overlapping outputs)
	for idx, line := range lines {
		var entry logger.LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is corrupted: %v; Raw line: %q", idx, err, line)
		}
	}
}

