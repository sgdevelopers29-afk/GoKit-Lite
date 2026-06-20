# Architecture — How GoKit-Lite Is Structured

This document explains the overall package design of GoKit-Lite, how the three
completed packages (`response`, `validator`, `auth`) fit together, and how a
typical HTTP request flows through them.

---

## Table of Contents

1. [Design Philosophy](#design-philosophy)
2. [Package Overview](#package-overview)
3. [Folder Structure](#folder-structure)
4. [Package Interaction Diagram](#package-interaction-diagram)
5. [Request Lifecycle](#request-lifecycle)
   - [Public Endpoint (Register / Login)](#public-endpoint-register--login)
   - [Protected Endpoint (/me)](#protected-endpoint-me)
6. [How the Packages Interact](#how-the-packages-interact)
7. [Dependency Graph](#dependency-graph)
8. [Adding a New Package](#adding-a-new-package)

---

## Design Philosophy

GoKit-Lite follows three core principles:

1. **Opt-in, not all-in.** Every package is a standalone Go sub-module import.
   You can use `response` without `auth`, or `validator` without `response`.
   There are no mandatory peer dependencies between packages.

2. **Standard library first.** The only external dependency is `golang-jwt/jwt`
   (required by `auth`). Everything else uses the Go standard library, keeping
   the dependency surface minimal and upgrade friction low.

3. **Composition over coupling.** Packages communicate through standard Go types
   — `error`, `http.Handler`, `context.Context` — not through internal shared
   state. This makes each package independently testable and replaceable.

---

## Package Overview

```
github.com/sgdevelopers29-afk/GoKit-Lite
│
├── response/    — JSON response envelope (Success / Error)
├── validator/   — tag-based struct validation
├── auth/        — JWT generation, validation, and HTTP middleware
│
├── examples/
│   └── user-api/    — complete runnable example combining all three packages
│
└── docs/        — documentation (this directory)
```

---

## Folder Structure

```
GoKit-Lite/
│
├── auth/
│   ├── auth.go          — Manager, GenerateToken, ValidateToken, package-level wrappers
│   ├── claims.go        — Claims struct (UserID, Email, Role + jwt.RegisteredClaims)
│   ├── errors.go        — sentinel error variables
│   ├── middleware.go    — RequireAuth middleware, ClaimsFromContext
│   ├── auth_test.go     — unit tests
│   └── example_test.go  — runnable Go doc examples
│
├── response/
│   ├── response.go      — Response struct, Success(), Error()
│   └── response_test.go — unit tests
│
├── validator/
│   ├── validator.go     — Validate, ValidateAll, all built-in rules, custom registration
│   ├── errors.go        — ValidationError, Result
│   ├── validator_test.go — unit tests
│   └── example_test.go  — runnable Go doc examples
│
├── examples/
│   └── user-api/
│       ├── main.go      — server bootstrap
│       ├── routes.go    — route table and middleware wiring
│       ├── handlers.go  — HTTP handler functions
│       ├── models.go    — request / response structs with validator tags
│       └── README.md
│
├── docs/
│   ├── getting-started.md
│   ├── response.md
│   ├── validator.md
│   ├── auth.md
│   ├── architecture.md  — this file
│   └── roadmap.md
│
├── go.mod
├── go.sum
└── README.md
```

---

## Package Interaction Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                        Your Application                      │
│                                                              │
│  ┌──────────┐    ┌──────────────┐    ┌────────────────────┐ │
│  │ response │    │  validator   │    │       auth         │ │
│  │          │    │              │    │                    │ │
│  │ Success()│    │ Validate()   │    │ GenerateToken()    │ │
│  │ Error()  │    │ ValidateAll()│    │ ValidateToken()    │ │
│  │          │    │ Register()   │    │ RequireAuth()      │ │
│  │ Response │    │ Result       │    │ ClaimsFromContext()│ │
│  └────┬─────┘    └──────┬───────┘    └────────┬───────────┘ │
│       │                 │                     │             │
│       └────────── used by handlers ───────────┘             │
└──────────────────────────────────────────────────────────────┘
```

Each package has a **narrow, focused interface**. No package imports any other
GoKit-Lite package — they interact only through the types defined by the Go
standard library.

---

## Request Lifecycle

### Public Endpoint (Register / Login)

```
Client
  │
  │  POST /register
  │  { "name":"Ganesh","email":"ganesh@gmail.com","password":"123456" }
  │
  ▼
net/http ServeMux
  │
  ▼
routes.go ──► handler function (handleRegister)
  │
  │  1. json.Decode(r.Body) → RegisterRequest struct
  │
  ▼
validator.ValidateAll(req)
  │  Checks all fields in one pass:
  │    Name      → required, minLength, maxLength
  │    Email     → required, email
  │    Password  → required, minLength
  │
  ├──► [validation failed]
  │         │
  │         ▼
  │    response.Error("validation failed")
  │    JSON: { "success":false, "message":"validation failed" }
  │    HTTP 422 ──────────────────────────────────────────► Client
  │
  └──► [all valid]
            │
            ▼
       auth.GenerateToken(Claims{...})
            │  Signs a JWT with HMAC-SHA256
            │  Sets IssuedAt and ExpiresAt automatically
            │
            ├──► [error]
            │        ▼
            │   response.Error("could not generate token")
            │   HTTP 500 ───────────────────────────────► Client
            │
            └──► [success]
                      │
                      ▼
                 response.Success(AuthPayload{User, Token})
                 JSON: { "success":true, "message":"success",
                         "data":{ "user":{...}, "token":"eyJ..." }}
                 HTTP 201 ─────────────────────────────► Client
```

---

### Protected Endpoint (/me)

```
Client
  │
  │  GET /me
  │  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  │
  ▼
net/http ServeMux
  │
  ▼
auth.RequireAuth(next)          ← middleware layer
  │
  │  1. Reads Authorization header
  │  2. Validates "Bearer <token>" format
  │  3. Calls auth.ValidateToken(tokenString)
  │       - Verifies HMAC-SHA256 signature
  │       - Checks expiration (exp claim)
  │
  ├──► [invalid / expired / missing]
  │         │
  │         ▼
  │    HTTP 401 Unauthorized ──────────────────────────► Client
  │
  └──► [valid token]
            │
            │  Injects *Claims into r.Context()
            │
            ▼
       handleMe(w, r)
            │
            │  auth.ClaimsFromContext(r.Context()) → *Claims
            │
            ▼
       response.Success(ProfileResponse{...})
       JSON: { "success":true, "message":"success",
               "data":{ "user_id":"usr_001","email":"...","role":"user" }}
       HTTP 200 ─────────────────────────────────────► Client
```

---

## How the Packages Interact

Although the packages are decoupled, they compose naturally in a handler:

```go
// handlers.go — typical handler combining all three packages
func handleRegister(w http.ResponseWriter, r *http.Request) {

    // ── validator: decode & check all fields ──────────────────────────────────
    var req RegisterRequest               // struct with validator tags
    json.NewDecoder(r.Body).Decode(&req)

    result := validator.ValidateAll(req)  // collect every broken rule
    if !result.Valid {
        writeJSON(w, 422, response.Error("validation failed"))
        return
    }

    // ── auth: issue a JWT ─────────────────────────────────────────────────────
    token, err := auth.GenerateToken(auth.Claims{
        UserID: "usr_001",
        Email:  req.Email,
        Role:   "user",
    })
    if err != nil {
        writeJSON(w, 500, response.Error("token error"))
        return
    }

    // ── response: uniform success envelope ────────────────────────────────────
    writeJSON(w, 201, response.Success(AuthPayload{Token: token}))
}
```

The three packages each own a distinct responsibility:

| Package | Owns | Does NOT own |
|---|---|---|
| `validator` | Field-level business rules | HTTP, responses, tokens |
| `auth` | Token generation / validation / middleware | Validation rules, response format |
| `response` | JSON envelope shape | Validation logic, tokens, HTTP status codes |

---

## Dependency Graph

```
response     (no GoKit-Lite dependencies)
     │
     └── used by: handlers, auth middleware

validator    (no GoKit-Lite dependencies)
     │
     └── used by: handlers

auth         (depends on: golang-jwt/jwt/v5)
     │
     └── used by: handlers, routes (RequireAuth middleware)

examples/user-api
     ├── imports response
     ├── imports validator
     └── imports auth
```

No circular dependencies. No shared global state between packages (auth has a
package-level `defaultManager` but it is self-contained).

---

## Adding a New Package

To add a new package to GoKit-Lite (e.g. `cache`):

1. Create a new directory: `mkdir cache`
2. Add `package cache` at the top of every `.go` file.
3. The module path for importing is automatically
   `github.com/sgdevelopers29-afk/GoKit-Lite/cache` — no changes to `go.mod`
   are needed because it's the same module.
4. Write tests in `cache/cache_test.go`.
5. Add runnable doc examples in `cache/example_test.go`.
6. Add documentation to `docs/cache.md`.
7. Open a PR targeting `develop`.
