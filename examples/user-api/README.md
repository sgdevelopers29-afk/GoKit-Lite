# user-api — GoKit-Lite Example

A complete, runnable REST API example that demonstrates how to use
**response**, **validator**, and **auth** from
[GoKit-Lite](https://github.com/sgdevelopers29-afk/GoKit-Lite) together in a
realistic application.

---

## What This Example Covers

| Feature | Package | Where |
|---|---|---|
| Standardised JSON responses | `response` | Every handler |
| Full-payload validation (all errors) | `validator.ValidateAll` | `handleRegister` |
| Fail-fast validation | `validator.Validate` | `handleLogin` |
| JWT generation | `auth.GenerateToken` | Register & Login handlers |
| JWT validation middleware | `auth.RequireAuth` | `GET /me` route |
| Claims extraction from context | `auth.ClaimsFromContext` | `handleMe` |

---

## Project Structure

```
examples/user-api/
├── main.go       — server bootstrap, auth configuration
├── routes.go     — route registration, middleware wiring
├── handlers.go   — HTTP handler functions
├── models.go     — request/response structs with validator tags
└── README.md     — this file
```

---

## Prerequisites

- **Go 1.21+** installed ([download](https://go.dev/dl/))
- The GoKit-Lite module cloned or imported

---

## Running the Example

```bash
# From the repository root:
cd examples/user-api
go run .
```

The server starts on **http://localhost:8080** and prints:

```
╔══════════════════════════════════════════════════╗
║         GoKit-Lite — User API Example            ║
╠══════════════════════════════════════════════════╣
║  Listening on  http://localhost:8080             ║
║                                                  ║
║  Routes:                                         ║
║    GET  /health      — liveness probe            ║
║    POST /register    — create account            ║
║    POST /login       — authenticate              ║
║    GET  /me          — profile (auth required)   ║
╚══════════════════════════════════════════════════╝
```

---

## API Reference & curl Examples

### GET /health
Check that the server is alive.

```bash
curl http://localhost:8080/health
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "success",
  "data": { "status": "ok" }
}
```

---

### POST /register
Create a new user account. All fields are validated with `validator.ValidateAll`
so every broken rule is reported in a single round-trip.

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ganesh","email":"ganesh@gmail.com","password":"123456"}'
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "user": {
      "id": "usr_gane",
      "name": "Ganesh",
      "email": "ganesh@gmail.com",
      "role": "user"
    },
    "token": "<jwt-token>"
  }
}
```

**Validation failure (422 Unprocessable Entity):**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"","email":"not-an-email","password":"123"}'
```

```json
{
  "success": false,
  "message": "validation failed"
}
```
_(Validation details are also logged to stdout by the server.)_

---

### POST /login
Authenticate with email and password.

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ganesh@gmail.com","password":"123456"}'
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "user": { "id": "usr_gane", "email": "ganesh@gmail.com", "role": "user" },
    "token": "<jwt-token>"
  }
}
```

---

### GET /me _(requires JWT)_
Return the authenticated user's profile read directly from the JWT claims.

```bash
# Copy the token from the login response and use it here:
TOKEN="<paste-token-here>"

curl http://localhost:8080/me \
  -H "Authorization: Bearer $TOKEN"
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "user_id": "usr_gane",
    "email": "ganesh@gmail.com",
    "role": "user"
  }
}
```

**Without a token (401 Unauthorized):**
```json
auth: missing Authorization header
```

---

## How the Packages Work Together

```
HTTP Request
     │
     ▼
  routes.go            ← decides which handler and whether auth is required
     │
     ├─ (public)──────► handler
     │                     │
     │                   json.Decode
     │                     │
     │                  validator.Validate / validator.ValidateAll
     │                     │  (returns all field errors at once)
     │                     │
     │                  auth.GenerateToken
     │                     │  (signs a JWT with the configured secret)
     │                     │
     │                  response.Success / response.Error
     │                     │  (uniform JSON envelope)
     │                     ▼
     │                  Client
     │
     └─ (protected)───► auth.RequireAuth (middleware)
                           │  validates JWT, injects *Claims into context
                           │
                           ▼
                        handler
                           │
                        auth.ClaimsFromContext
                           │
                        response.Success
                           │
                           ▼
                        Client
```

---

## Key Learning Points

1. **`validator.ValidateAll`** — use during registration so users see *all*
   broken fields at once, not just the first one.

2. **`validator.Validate`** — use for login or any operation where you want
   to bail out on the first invalid field (fail-fast).

3. **`auth.SetSecret`** must be called *before* any `GenerateToken` or
   `ValidateToken` call.  In production, read the secret from an environment
   variable.

4. **`auth.RequireAuth`** is a standard `net/http` middleware — wrap any
   `http.Handler` with it to protect a route.

5. **`response.Success` / `response.Error`** produce a consistent
   `{"success":bool,"message":"...","data":...}` envelope that all your
   API clients can rely on.

---

## Running the Tests

From the repository root:

```bash
go test ./...
```

Or target specific packages:

```bash
go test ./validator/... -v
go test ./auth/...     -v
go test ./response/... -v
```
