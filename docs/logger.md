# Logger Package (`logger/`)

The `logger` package provides a simple, lightweight, and structured JSON logging mechanism built entirely on the Go standard library, avoiding heavy third-party dependencies.

## Features (v0.1)
- Outputs logs in a structured JSON format natively parseable by log aggregation tools (e.g., Datadog, Elasticsearch).
- Supports informational (`INFO`) and error (`ERROR`) log levels.
- Automatically appends a standard RFC3339 UTC timestamp to all entries.

## API

```go
// Info logs an informational message.
func Info(message string)

// Error logs an error message.
func Error(message string)
```

## Log Format

```json
{
  "level": "INFO",
  "message": "server started",
  "timestamp": "2026-06-12T15:04:05Z"
}
```

## Example Usage

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/logger"

func main() {
    logger.Info("application initialized")
    
    // Simulating an error
    err := doSomething()
    if err != nil {
        logger.Error("failed to perform action: " + err.Error())
    }
}
```
