// Package logger provides simple, structured JSON logging capabilities for GoKit-Lite.
package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

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
		fmt.Printf(`{"level":"%s","message":"%s","error":"failed to serialize json"}`+"\n", level, message)
		return
	}

	// Output the JSON string
	fmt.Println(string(bytes))
}

// Info logs an informational message in JSON format.
func Info(message string) {
	log("INFO", message)
}

// Error logs an error message in JSON format.
func Error(message string) {
	log("ERROR", message)
}
