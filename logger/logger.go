// Package logger provides simple, structured JSON logging capabilities for GoKit-Lite.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

var out io.Writer = os.Stdout

// SetOutput allows redirecting the JSON logs from standard out to any io.Writer.
func SetOutput(w io.Writer) {
	out = w
}

// LogEntry represents a single structured log message.
type LogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// log is an internal helper function that constructs and prints a JSON log entry.
func log(level string, message string) {
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal the struct into a JSON byte slice
	bytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to a basic format if JSON marshaling fails (unlikely here)
		fmt.Fprintf(out, `{"level":"%s","message":"%s","error":"failed to serialize json"}`+"\n", level, message)
		return
	}

	// Output the JSON string
	fmt.Fprintln(out, string(bytes))
}

// Info logs an informational message in JSON format.
func Info(message string) {
	log("INFO", message)
}

// Infof formats and logs an informational message.
func Infof(format string, args ...any) {
	log("INFO", fmt.Sprintf(format, args...))
}

// Error logs an error message in JSON format.
func Error(message string) {
	log("ERROR", message)
}

// Errorf formats and logs an error message.
func Errorf(format string, args ...any) {
	log("ERROR", fmt.Sprintf(format, args...))
}

// Warnf formats and logs a warning message.
func Warnf(format string, args ...any) {
	log("WARN", fmt.Sprintf(format, args...))
}

// Debugf formats and logs a debug message.
func Debugf(format string, args ...any) {
	log("DEBUG", fmt.Sprintf(format, args...))
}
