# Getting Started with GoKit-Lite

> **GoKit-Lite** is a lightweight, modular toolkit for building scalable Go
> backend applications. It gives you production-ready utilities for the most
> common tasks — validation, authentication, and standardised API responses —
> so you can focus on your business logic instead of boilerplate.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Project Overview](#project-overview)
4. [Supported Packages](#supported-packages)
5. [Quick Start](#quick-start)
6. [Running the Full Example](#running-the-full-example)
7. [Running the Tests](#running-the-tests)
8. [Next Steps](#next-steps)

---

## Prerequisites

- **Go 1.21 or later** — [download here](https://go.dev/dl/)
- Basic familiarity with Go modules and the `net/http` package

---

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

GoKit-Lite uses Go modules. Run the command above inside your module directory
(where your `go.mod` lives). The only external dependency it brings in is
[`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) — everything else
relies on the Go standard library.

---

## Project Overview

GoKit-Lite is organised as a collection of independent, opt-in packages that
live under the same module path. You only import what you need:

```
github.com/sgdevelopers29-afk/GoKit-Lite/response
github.com/sgdevelopers29-afk/GoKit-Lite/validator
github.com/sgdevelopers29-afk/GoKit-Lite/auth
```

Each package has zero awareness of the others — they compose cleanly but are
completely decoupled. This design means you can adopt one package without
pulling in the rest.

---

## Supported Packages

| Package | Purpose | Status |
|---|---|---|
| `response` | Standardised `{"success":bool,"message":"...","data":...}` JSON envelope | ✅ Stable |
| `validator` | Tag-based struct validation with 9 built-in rules + custom validators | ✅ Stable |
| `auth` | JWT generation, validation, and `net/http` middleware | ✅ Stable |
| `cache` | In-memory / distributed caching | 🔜 Planned |
| `logger` | Structured JSON logging | 🔜 Planned |
| `monitor` | Metrics and health-check utilities | 🔜 Planned |
| `ratelimit` | Token-bucket / sliding-window rate limiting | 🔜 Planned |

---

## Quick Start

### 1. Standardised API Responses

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/sgdevelopers29-afk/GoKit-Lite/response"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
    // Success response — wraps any payload
    resp := response.Success(map[string]string{"greeting": "Hello, World!"})
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
    // Output: {"success":true,"message":"success","data":{"greeting":"Hello, World!"}}

    // Error response — carries a human-readable message
    errResp := response.Error("user not found")
    json.NewEncoder(w).Encode(errResp)
    // Output: {"success":false,"message":"user not found"}
}
```

---

### 2. Struct Validation

```go
package main

import (
    "fmt"

    "github.com/sgdevelopers29-afk/GoKit-Lite/validator"
)

type RegisterRequest struct {
    Name     string `required:"true" minLength:"2" maxLength:"50"`
    Email    string `required:"true" email:"true"`
    Password string `required:"true" minLength:"6"`
}

func main() {
    req := RegisterRequest{
        Name:     "G",         // too short (minLength:"2" allows ≥ 2 chars)
        Email:    "bad-email",
        Password: "123",       // too short (minLength:"6")
    }

    // Validate — fail-fast: returns the first broken rule only.
    if err := validator.Validate(req); err != nil {
        fmt.Println(err) // field Name failed rule "minLength": ...
    }

    // ValidateAll — collects every broken rule across all fields.
    result := validator.ValidateAll(req)
    if !result.Valid {
        for _, e := range result.Errors {
            fmt.Printf("Field: %-10s Rule: %s\n", e.Field, e.Rule)
        }
    }
}
```

---

### 3. JWT Authentication

```go
package main

import (
    "fmt"
    "time"

    "github.com/sgdevelopers29-afk/GoKit-Lite/auth"
)

func main() {
    // Configure once at startup — read secret from environment in production!
    auth.SetSecret("my-super-secret-key")
    auth.SetTokenDuration(24 * time.Hour)

    // Generate a token
    token, err := auth.GenerateToken(auth.Claims{
        UserID: "usr_001",
        Email:  "alice@example.com",
        Role:   "admin",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Token:", token)

    // Validate a token
    claims, err := auth.ValidateToken(token)
    if err != nil {
        panic(err)
    }
    fmt.Println("UserID:", claims.UserID)
    fmt.Println("Email:", claims.Email)
    fmt.Println("Role:", claims.Role)
}
```

---

### 4. Protecting Routes with Middleware

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/sgdevelopers29-afk/GoKit-Lite/auth"
)

func main() {
    auth.SetSecret("my-super-secret-key")

    mux := http.NewServeMux()

    // Protected handler
    protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims, _ := auth.ClaimsFromContext(r.Context())
        fmt.Fprintf(w, "Hello, %s!", claims.UserID)
    })

    // Wrap with RequireAuth — automatically validates Bearer token
    mux.Handle("/dashboard", auth.RequireAuth(protectedHandler))

    http.ListenAndServe(":8080", mux)
}
```

---

## Running the Full Example

The repository ships with a complete, working User API example that combines
all three packages:

```bash
# Clone the repository
git clone https://github.com/sgdevelopers29-afk/GoKit-Lite.git
cd GoKit-Lite

# Run the example server
cd examples/user-api
go run .
```

The server starts at **http://localhost:8080**. See
[`examples/user-api/README.md`](../examples/user-api/README.md) for a full list
of curl commands to exercise every endpoint.

---

## Running the Tests

```bash
# Run the entire test suite from the repository root
go test ./...

# Run with verbose output
go test ./... -v

# Run a single package
go test ./validator/... -v
go test ./auth/...      -v
go test ./response/...  -v
```

---

## Next Steps

| Goal | Where to look |
|---|---|
| Understand the `response` package | [docs/response.md](response.md) |
| Learn every validation rule | [docs/validator.md](validator.md) |
| Deep-dive into JWT auth | [docs/auth.md](auth.md) |
| See how packages fit together | [docs/architecture.md](architecture.md) |
| Track future features | [docs/roadmap.md](roadmap.md) |
