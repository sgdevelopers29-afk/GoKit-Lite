// Package logger provides simple, structured JSON logging capabilities for GoKit-Lite.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the severity threshold of a log message.
type Level int

// Log level constants.
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the uppercase string representation of the Level.
func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string representation of a log level into a Level constant.
// It is case-insensitive and trims white space. If unrecognized, it defaults to LevelInfo.
func ParseLevel(lvl string) Level {
	switch strings.ToUpper(strings.TrimSpace(lvl)) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// LogEntry represents a single structured log message.
type LogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

var (
	mu           sync.RWMutex
	out          io.Writer = os.Stdout
	currentLevel Level     = LevelInfo
)

func init() {
	if lvlEnv := os.Getenv("LOG_LEVEL"); lvlEnv != "" {
		currentLevel = ParseLevel(lvlEnv)
	}
}

// SetLevel dynamically changes the active log level threshold.
func SetLevel(lvl Level) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = lvl
}

// GetLevel returns the current active log level threshold.
func GetLevel() Level {
	mu.RLock()
	defer mu.RUnlock()
	return currentLevel
}

// SetOutput changes the output destination of the logger.
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	out = w
}

// SetOutputFile directs the log output to a file at the specified path.
func SetOutputFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("logger: failed to open log file %q: %w", path, err)
	}
	SetOutput(file)
	return nil
}

// log is an internal helper function that constructs and prints a JSON log entry.
func log(level Level, message string) {
	mu.RLock()
	activeLevel := currentLevel
	writer := out
	mu.RUnlock()

	if level < activeLevel {
		return
	}

	entry := LogEntry{
		Level:     level.String(),
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal the struct into a JSON byte slice
	bytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to a basic format if JSON marshaling fails (unlikely here)
		fallbackMsg := fmt.Sprintf(`{"level":"%s","message":"%s","error":"failed to serialize json"}`+"\n", level.String(), message)
		mu.Lock()
		_, _ = fmt.Fprint(writer, fallbackMsg)
		mu.Unlock()
		return
	}

	// Output the JSON string with a trailing newline
	mu.Lock()
	_, _ = writer.Write(append(bytes, '\n'))
	mu.Unlock()
}

// Trace logs a trace-level message in JSON format.
func Trace(message string) {
	log(LevelTrace, message)
}

// Debug logs a debug-level message in JSON format.
func Debug(message string) {
	log(LevelDebug, message)
}

// Info logs an informational message in JSON format.
func Info(message string) {
	log(LevelInfo, message)
}

// Warn logs a warning-level message in JSON format.
func Warn(message string) {
	log(LevelWarn, message)
}

// Error logs an error message in JSON format.
func Error(message string) {
	log(LevelError, message)
}

