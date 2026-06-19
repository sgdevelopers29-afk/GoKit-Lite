# `auth` — JWT Authentication

The `auth` package provides a complete JWT (JSON Web Token) workflow for Go HTTP
services: generate signed tokens on login, validate them on every protected
request, and extract the user's identity from the request context — all with the
standard `net/http` interface and zero additional dependencies beyond
`golang-jwt/jwt/v5`.

---

## Table of Contents

1. [What is a JWT?](#what-is-a-jwt)
2. [Installation](#installation)
3. [The Claims Struct](#the-claims-struct)
4. [Configuration](#configuration)
   - [SetSecret](#setsecret)
   - [SetTokenDuration](#settokenduration)
5. [Functions](#functions)
   - [GenerateToken](#generatetoken)
   - [ValidateToken](#validatetoken)
6. [Middleware](#middleware)
   - [RequireAuth](#requireauth)
   - [ClaimsFromContext](#claimsfromcontext)
7. [Manager API (Advanced)](#manager-api-advanced)
8. [Error Reference](#error-reference)
9. [Complete Example](#complete-example)
10. [Token Expiration](#token-expiration)
11. [Secret Configuration](#secret-configuration)
12. [Common Mistakes](#common-mistakes)

---

## What is a JWT?

A **JSON Web Token (JWT)** is a compact, URL-safe string used to represent
claims between two parties. In a typical web API:

1. The server **generates** a JWT after verifying the user's credentials.
2. The server **returns** the JWT to the client.
3. On every subsequent request, the client **sends** the JWT in the
   `Authorization: Bearer <token>` header.
4. The server **validates** the JWT's signature and expiration to authenticate
   the request — no database lookup needed.

A JWT has three parts separated by dots (`header.payload.signature`):

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
.eyJ1c2VyX2lkIjoidXNyXzAwMSIsImVtYWlsIjoiYWxpY2VAZXhhbXBsZS5jb20ifQ
.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

GoKit-Lite uses HMAC-SHA256 (`HS256`) for signing — the same secret signs and
verifies the token, so keep it private.

---

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/auth"
```

---

## The Claims Struct

```go
type Claims struct {
    UserID string `json:"user_id,omitempty"`
    Email  string `json:"email,omitempty"`
    Role   string `json:"role,omitempty"`
    jwt.RegisteredClaims
}
```

| Field | JSON key | Description |
|---|---|---|
| `UserID` | `"user_id"` | Uniquely identifies the authenticated user |
| `Email` | `"email"` | The user's email address |
| `Role` | `"role"` | The user's role (e.g. `"admin"`, `"user"`) |
| `RegisteredClaims` | (embedded) | Standard JWT fields: `iss`, `sub`, `aud`, `exp`, `nbf`, `iat`, `jti` |

The embedded `jwt.RegisteredClaims` means your `Claims` automatically handles
the standard expiration (`exp`) and issued-at (`iat`) fields that GoKit-Lite
populates for you.

You only need to fill in the custom fields (`UserID`, `Email`, `Role`) — the
rest is automatic.

---

## Configuration

### `SetSecret`

```go
func SetSecret(secret string)
```

Sets the HMAC-SHA256 signing key used by the **default package-level manager**.
This must be called before `GenerateToken` or `ValidateToken`.

```go
auth.SetSecret("my-secret-key")
```

> ⚠️ **In production, always read the secret from an environment variable or
> secrets manager. Never commit a secret to version control.**

---

### `SetTokenDuration`

```go
func SetTokenDuration(d time.Duration)
```

Sets the lifetime of tokens issued by the default manager. If not called, the
default is **24 hours**.

```go
auth.SetTokenDuration(7 * 24 * time.Hour) // 7-day tokens
auth.SetTokenDuration(15 * time.Minute)    // short-lived access tokens
```

---

## Functions

### `GenerateToken`

```go
func GenerateToken(claims Claims) (string, error)
```

Creates a signed JWT string from the provided `Claims`. If `IssuedAt` or
`ExpiresAt` are not already set in the claims, `GenerateToken` populates them
automatically using the configured duration.

**Returns:** the signed JWT string, or an error (see [Error Reference](#error-reference)).

```go
auth.SetSecret("super-secret")
auth.SetTokenDuration(24 * time.Hour)

token, err := auth.GenerateToken(auth.Claims{
    UserID: "usr_001",
    Email:  "alice@example.com",
    Role:   "admin",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(token)
// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

### `ValidateToken`

```go
func ValidateToken(tokenString string) (*Claims, error)
```

Parses a JWT string, verifies its HMAC-SHA256 signature, and checks that the
token has not expired. Returns the extracted `*Claims` on success.

**Returns:** `*Claims` on success, or one of the sentinel errors below.

```go
claims, err := auth.ValidateToken(tokenString)
if err != nil {
    switch err {
    case auth.ErrExpiredToken:
        // Token expired — ask the user to log in again.
    case auth.ErrInvalidSignature:
        // Token has been tampered with.
    default:
        // Generic invalid token.
    }
}

fmt.Println(claims.UserID) // "usr_001"
fmt.Println(claims.Role)   // "admin"
```

---

## Middleware

### `RequireAuth`

```go
func RequireAuth(next http.Handler) http.Handler
```

`RequireAuth` is a standard `net/http` middleware. Wrap any `http.Handler` with
it to make that route protected.

**What it does:**

1. Reads the `Authorization` header.
2. Validates that it has the form `Bearer <token>`.
3. Calls `ValidateToken` on the token string.
4. On success: injects `*Claims` into `r.Context()` and calls `next`.
5. On failure: responds with **401 Unauthorized** and stops the chain.

```go
mux := http.NewServeMux()

// Public route — no token needed
mux.HandleFunc("/login", handleLogin)

// Protected route — RequireAuth validates the JWT before the handler runs
mux.Handle("/dashboard", auth.RequireAuth(
    http.HandlerFunc(handleDashboard),
))

http.ListenAndServe(":8080", mux)
```

**Authorization header format:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

### `ClaimsFromContext`

```go
func ClaimsFromContext(ctx context.Context) (*Claims, bool)
```

Extracts the `*Claims` that `RequireAuth` injected into the request context.
Returns `nil, false` if no claims are present (i.e. `RequireAuth` was not used
on the route).

```go
func handleDashboard(w http.ResponseWriter, r *http.Request) {
    claims, ok := auth.ClaimsFromContext(r.Context())
    if !ok {
        http.Error(w, "no auth claims", http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "Welcome, %s (%s)!", claims.Email, claims.Role)
}
```

---

## Manager API (Advanced)

The package-level functions (`SetSecret`, `GenerateToken`, etc.) are thin
wrappers around a shared `*Manager`. For applications that need **multiple
independent token configurations** (e.g. access tokens vs. refresh tokens with
different secrets and lifetimes), create a dedicated `Manager`:

```go
// Access token manager — short lifetime
accessManager := auth.NewManager("access-secret", 15*time.Minute)

// Refresh token manager — long lifetime
refreshManager := auth.NewManager("refresh-secret", 30*24*time.Hour)

// Generate tokens with each manager
accessToken, _ := accessManager.GenerateToken(auth.Claims{UserID: "usr_1"})
refreshToken, _ := refreshManager.GenerateToken(auth.Claims{UserID: "usr_1"})

// Protect a route with a specific manager's middleware
mux.Handle("/api/data", accessManager.RequireAuth(
    http.HandlerFunc(handleAPIData),
))
```

`Manager` exposes the same surface as the package-level API:

```go
m.SetSecret(secret)
m.SetTokenDuration(d)
m.GenerateToken(claims) (string, error)
m.ValidateToken(tokenString) (*Claims, error)
m.RequireAuth(next http.Handler) http.Handler
```

---

## Error Reference

| Sentinel | Meaning | When you see it |
|---|---|---|
| `auth.ErrMissingSecret` | No signing key has been set | Called `GenerateToken`/`ValidateToken` before `SetSecret` |
| `auth.ErrExpiredToken` | The `exp` claim is in the past | Token issued earlier than `tokenDuration` ago |
| `auth.ErrInvalidSignature` | HMAC signature does not match | Token was tampered with or signed with a different secret |
| `auth.ErrInvalidToken` | Token is malformed or unparseable | Garbage string, truncated token, wrong format |
| `auth.ErrInvalidClaims` | Claims could not be cast to `*Claims` | Should never happen with standard usage |
| `auth.ErrMissingAuthHeader` | `Authorization` header absent | Client did not send the header |
| `auth.ErrInvalidAuthHeader` | Header is not `Bearer <token>` | Client sent a malformed header |

---

## Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/sgdevelopers29-afk/GoKit-Lite/auth"
    "github.com/sgdevelopers29-afk/GoKit-Lite/response"
)

func main() {
    // 1. Configure once at startup
    auth.SetSecret("change-me-in-production")
    auth.SetTokenDuration(24 * time.Hour)

    mux := http.NewServeMux()

    // 2. Public: issue tokens
    mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        // (In a real app, verify credentials here)
        token, err := auth.GenerateToken(auth.Claims{
            UserID: "usr_42",
            Email:  "bob@example.com",
            Role:   "user",
        })
        if err != nil {
            json.NewEncoder(w).Encode(response.Error(err.Error()))
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response.Success(map[string]string{"token": token}))
    })

    // 3. Protected: require a valid JWT
    mux.Handle("/profile", auth.RequireAuth(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            claims, _ := auth.ClaimsFromContext(r.Context())
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(response.Success(map[string]string{
                "user_id": claims.UserID,
                "email":   claims.Email,
                "role":    claims.Role,
            }))
        },
    )))

    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", mux)
}
```

---

## Token Expiration

When `GenerateToken` is called without explicit `ExpiresAt` or `IssuedAt`
values in the `Claims`, the package sets them automatically:

```
IssuedAt  = time.Now()
ExpiresAt = time.Now() + tokenDuration   (default: 24h)
```

To override, set the fields manually:

```go
import "github.com/golang-jwt/jwt/v5"

claims := auth.Claims{
    UserID: "usr_001",
    RegisteredClaims: jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
    },
}
token, _ := auth.GenerateToken(claims)
```

---

## Secret Configuration

| Environment | Recommended approach |
|---|---|
| Development | Hard-code in `main.go` for convenience |
| CI / Testing | Set as a CI environment secret |
| Staging / Production | Read from `os.Getenv("JWT_SECRET")` or a secrets manager (Vault, AWS Secrets Manager, etc.) |

**Production-ready secret setup:**

```go
import "os"

func main() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    auth.SetSecret(secret)
}
```

---

## Common Mistakes

### ❌ Forgetting to call `SetSecret`

```go
// Wrong — no secret set
token, err := auth.GenerateToken(claims)
// err = "auth: missing secret key"
```

```go
// Correct
auth.SetSecret("my-secret")
token, err := auth.GenerateToken(claims)
```

---

### ❌ Comparing errors with `==` instead of `errors.Is`

```go
// Wrong
if err == auth.ErrExpiredToken { ... }

// Correct
if errors.Is(err, auth.ErrExpiredToken) { ... }
```

---

### ❌ Reading claims without checking the `ok` flag

```go
// Wrong — may panic if RequireAuth was not applied
claims, _ := auth.ClaimsFromContext(r.Context())
fmt.Println(claims.UserID) // nil pointer dereference!

// Correct
claims, ok := auth.ClaimsFromContext(r.Context())
if !ok {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
}
```

---

### ❌ Putting sensitive data in the payload

JWTs are **signed, not encrypted**. Anyone can Base64-decode the payload and
read it. Never store passwords, raw credit-card numbers, or other secrets in the
`Claims`.

---

### ❌ Using the same secret for access and refresh tokens

If both token types share a secret, an attacker who obtains a refresh token
could craft a valid access token. Use separate `Manager` instances with
different secrets for access vs. refresh tokens.
