# GoKit-Lite

[![Go Reference](https://pkg.go.dev/badge/github.com/sgdevelopers29-afk/GoKit-Lite.svg)](https://pkg.go.dev/github.com/sgdevelopers29-afk/GoKit-Lite)
[![Go Report Card](https://goreportcard.com/badge/github.com/sgdevelopers29-afk/GoKit-Lite)](https://goreportcard.com/report/github.com/sgdevelopers29-afk/GoKit-Lite)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

GoKit-Lite is a lightweight, modular toolkit for building scalable Go applications. It provides essential, easy-to-use utilities for common backend tasks, ensuring consistency, clean code, and rapid development.

This is an open-source project, actively maintained and open to contributions.

## Features

* **Response:** Standardized API response format for Success and Error payloads.
* **Validator:** Fast, tag-based struct validation (e.g., `required:"true"`).
* **Config:** Centralized configuration and environment variable management.
* **Logger:** Simple, structured JSON logging for better observability.
* *(Coming Soon)* Cache, Rate Limiter, Auth, and Monitoring modules.

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

## Usage Examples

### API Responses
```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/response"

// Success response
resp := response.Success(map[string]string{"name": "Ganesh"})

// Error response
errResp := response.Error("User not found")
```

### Struct Validation
```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/validator"

type User struct {
	Name  string `required:"true"`
	Email string `required:"true"`
}

err := validator.Validate(User{Name: "Ganesh"}) 
// Returns: error "field Email is required"
```

### Structured Logging
```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/logger"

logger.Info("Server started on port 8080")
logger.Error("Database connection failed")
```

### Configuration
```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/config"

port := config.Get("PORT")
```

# Project Structure

```text
gokit-lite/
├── response/
├── validator/
├── config/
├── logger/
├── cache/
├── ratelimit/
├── auth/
├── monitor/
├── examples/
├── docs/
├── tests/
└── .github/workflows/
```

## Contributing

Thank you for considering contributing to this open-source project — contributions are welcome!

Please follow these guidelines to help us review and merge your changes quickly:

- **Branching:** Fork the repo and create a branch from `develop`, e.g. `feature/short-description` or `fix/short-description`.
- **Pull Requests:** Open PRs targeting `develop`. Include a clear description, related issue (if any), and steps to reproduce or test.
- **Commits:** Use clear, imperative commit messages (e.g., "Add cache middleware"). Reference issue IDs when applicable.
- **Tests:** Add or update unit tests for new behavior. Run `go test ./...` locally and ensure tests pass.
- **Code Style:** Format code with `gofmt` and run `go vet`. Prefer idiomatic Go and keep changes focused.
- **CI & Checks:** Ensure CI checks pass before requesting review. Maintainers may request changes; please address feedback promptly.
- **Small Changes:** Documentation, typo fixes, and small improvements are welcome via PRs.
- **License:** By contributing, you agree your contributions will be licensed under the project's MIT license.

If you're unsure where to start, check the `issues` tab for good first issues or open one to discuss your idea.
