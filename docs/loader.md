# Config Loader (`config/`)

The `config` loader (v2) extends the base configuration package by providing `.env` file parsing capabilities.

## Features (v2)
- Added `.env` file loading support via `Load(path string)`.
- Parses `KEY=VALUE` pairs directly into the system environment (`os.Setenv`).
- Automatically ignores empty lines and comments (lines starting with `#`).
- Safely trims whitespace around keys and values.

## API

```go
// Load reads an environment file from the given path and loads its contents into the system environment.
// It returns an error if the file cannot be read or if a line has an invalid format.
func Load(path string) error
```

## Example Usage

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/config"

func main() {
    // Load variables from a .env file
    if err := config.Load(".env"); err != nil {
        // Handle error
    }

    // Safely retrieve loaded variables
    port := config.Get("PORT")
    if port == "" {
        port = "8080"
    }
}
```
