# Config Package (`config/`)

The `config` package provides a foundational configuration management system. Currently, it acts as a centralized abstraction layer for environment variable retrieval.

## Features (v0.1)
- Retrieves environment variables via a safe `os.Getenv` wrapper.
- Establishes a clean architecture for future enhancements (e.g., struct mapping, default values).
- Provides a `.env.example` template for documenting required application settings.

## API

```go
// Get retrieves the value of the environment variable named by the key.
// It returns an empty string if the variable is not present.
func Get(key string) string
```

## Example Usage

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/config"

func main() {
    port := config.Get("PORT")
    if port == "" {
        port = "8080" // fallback
    }
}
```
