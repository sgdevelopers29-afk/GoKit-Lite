# GoKit-Lite Auth V1 Walkthrough

The Auth V1 package has been fully implemented according to your requirements. It focuses on clean, idiomatic Go architecture using `github.com/golang-jwt/jwt/v5`, ensuring safe open-source consumption with zero hard-dependencies on web frameworks like Gin or Fiber.

## 1. Branch & Dependencies

- Checked out from `develop` and created `feature/auth-v1`.
- Installed `github.com/golang-jwt/jwt/v5`.

## 2. JWT Architecture & Implementation

### Core JWT Concepts Used
- **HMAC (HS256)**: We use symmetric signing (`HS256`). This requires a single, secure secret key to both sign and validate the tokens.
- **Payload (`Claims`)**: The token payload uses custom JSON fields combined with standard JWT registered claims (like `exp` for expiration and `iat` for issued at).
- **Validation**: When parsing the token, the library verifies the signature matches the secret and inherently validates the expiration timestamps (`exp`).

### Package Structure

The package is split logically to separate concerns:

- **`auth/claims.go`**: Defines the `Claims` structure matching your `UserID`, `Email`, and `Role` format, embedding `jwt.RegisteredClaims` to handle standard JWT timestamps.
- **`auth/errors.go`**: Defines custom sentinel errors (e.g., `ErrInvalidToken`, `ErrExpiredToken`, `ErrMissingSecret`) making it easy for consumers to use `errors.Is(err, auth.ErrExpiredToken)`.
- **`auth/auth.go`**: Contains the core logic. Features a `Manager` struct for dependency injection workflows, paired with a global default instance to allow simple package-level calls like `auth.GenerateToken(...)` and `auth.SetSecret(...)`. 
- **`auth/middleware.go`**: Implements standard HTTP middleware and context extraction.
- **`auth/auth_test.go`**: Unit tests.
- **`auth/example_test.go`**: Executable GoDoc examples.

## 3. Middleware Design

> [!TIP]
> **Framework Agnosticism**
> The implemented middleware uses `net/http` standard types:
> `func RequireAuth(next http.Handler) http.Handler`
> 
> By using standard HTTP interfaces and `context.Context`, this middleware works natively with `net/http` routers (like Chi or standard `http.ServeMux`), while also being easily adaptable by Gin or Fiber developers using wrapper functions.

### Middleware Usage Example:
```go
protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    claims, _ := auth.ClaimsFromContext(r.Context())
    fmt.Fprintf(w, "Hello, user %s!", claims.UserID)
})

// Wrap handler with RequireAuth middleware
wrappedHandler := auth.RequireAuth(protectedHandler)
```

## 4. Test Coverage

Extensive tests were written targeting the core features:
- Missing secrets during generation/validation.
- Manipulated signing methods (protecting against "none" algorithm attacks).
- Token expiration boundaries.
- Invalid token strings and corrupted signatures.
- Context injection and extraction in the middleware.

**Test Results:**
```
PASS
coverage: 93.2% of statements
ok      github.com/sgdevelopers29-afk/GoKit-Lite/auth   1.865s  coverage: 93.2% of statements
```

## 5. Future Roadmap Context

The current `Claims` and `Manager` structures have been written linearly so that adding V2 features (like Refresh Tokens or Blacklisting via an interface) and V3 features (RBAC permission checks using the existing `Role` claim) will be natural, non-breaking additions to the existing API structure.
