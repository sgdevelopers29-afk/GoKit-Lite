# `response` — Standardised API Responses

The `response` package provides a single, consistent JSON envelope for every API
response your server sends. Instead of every handler inventing its own response
shape, you use `response.Success` or `response.Error` and every client gets the
same predictable structure.

---

## Table of Contents

1. [Installation](#installation)
2. [The Response Struct](#the-response-struct)
3. [Functions](#functions)
   - [Success](#success)
   - [Error](#error)
4. [Usage Examples](#usage-examples)
5. [Expected JSON Output](#expected-json-output)
6. [Common Use Cases](#common-use-cases)
7. [Best Practices](#best-practices)

---

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/response"
```

---

## The Response Struct

```go
type Response struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

| Field | Type | JSON key | Description |
|---|---|---|---|
| `Success` | `bool` | `"success"` | `true` on success, `false` on error |
| `Message` | `string` | `"message"` | Human-readable status (`"success"` or an error description) |
| `Data` | `any` | `"data"` | The response payload; omitted from JSON when `nil` |

The `Data` field uses the `omitempty` JSON tag — it is **absent from the JSON
output** when the value is `nil`. This keeps error responses compact.

---

## Functions

### `Success`

```go
func Success(data any) Response
```

`Success` constructs a `Response` with `Success: true`, `Message: "success"`,
and the provided `data` payload.

**Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `data` | `any` | Any value to embed as the response payload. Structs, maps, slices, and primitives are all valid. Pass `nil` to produce a bodyless success. |

**Returns:** A `Response` value ready to be JSON-encoded.

---

### `Error`

```go
func Error(message string) Response
```

`Error` constructs a `Response` with `Success: false`, the provided `message`,
and `Data: nil` (which is omitted from the JSON output).

**Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `message` | `string` | A human-readable description of what went wrong. |

**Returns:** A `Response` value ready to be JSON-encoded.

---

## Usage Examples

### Basic HTTP handler

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/sgdevelopers29-afk/GoKit-Lite/response"
)

type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

func getUser(w http.ResponseWriter, r *http.Request) {
    user := User{ID: "usr_001", Email: "alice@example.com"}

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response.Success(user))
}

func notFound(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    json.NewEncoder(w).Encode(response.Error("user not found"))
}
```

---

### Success with a struct payload

```go
type CreateUserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Token string `json:"token"`
}

resp := response.Success(CreateUserResponse{
    ID:    "usr_42",
    Name:  "Ganesh",
    Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
})
```

---

### Success with a plain map

```go
resp := response.Success(map[string]string{
    "status": "ok",
    "version": "1.0.0",
})
```

---

### Error response

```go
resp := response.Error("email address is already registered")
```

---

### Success with no data body

```go
// For operations like DELETE where there is nothing meaningful to return.
resp := response.Success(nil)
```

---

## Expected JSON Output

```go
response.Success(map[string]string{"greeting": "Hello!"})
```
```json
{
  "success": true,
  "message": "success",
  "data": { "greeting": "Hello!" }
}
```

---

```go
response.Error("invalid credentials")
```
```json
{
  "success": false,
  "message": "invalid credentials"
}
```
_(`"data"` is omitted because it is `nil` and the field uses `omitempty`.)_

---

```go
response.Success(nil)
```
```json
{
  "success": true,
  "message": "success"
}
```

---

## Common Use Cases

| Situation | Function | HTTP Status |
|---|---|---|
| Resource created | `response.Success(newResource)` | 201 Created |
| Resource fetched | `response.Success(resource)` | 200 OK |
| Resource deleted | `response.Success(nil)` | 200 OK |
| Validation failure | `response.Error("validation failed")` | 422 Unprocessable Entity |
| Not found | `response.Error("resource not found")` | 404 Not Found |
| Unauthorized | `response.Error("invalid or missing token")` | 401 Unauthorized |
| Internal server error | `response.Error("internal server error")` | 500 Internal Server Error |

---

## Best Practices

1. **Always set `Content-Type: application/json`** before calling
   `json.NewEncoder(w).Encode(...)`.

2. **Call `w.WriteHeader(statusCode)` before encoding** — once you write the
   body, Go's `http.ResponseWriter` locks the status to 200 if you haven't set
   it explicitly.

3. **Keep error messages user-safe.** Never expose raw stack traces or
   database error strings in `response.Error`. Log the internal error server-
   side and return a generic message.

4. **Use structured `data` payloads.** Prefer a named struct over a raw
   `map[string]any` so the JSON shape is documented and type-checked.

5. **Centralise JSON writing in a helper** to avoid repeating headers and
   status codes everywhere:

   ```go
   func writeJSON(w http.ResponseWriter, status int, v any) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(v)
   }

   // Usage:
   writeJSON(w, http.StatusOK,      response.Success(data))
   writeJSON(w, http.StatusNotFound, response.Error("not found"))
   ```
