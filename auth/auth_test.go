package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// This function test the basic JWT generation and validation
func TestGenerateAndValidateToken_Valid(t *testing.T) {
	mgr := NewManager("test-secret", 1*time.Hour)

	claims := Claims{
		UserID: "123",
		Email:  "test@example.com",
		Role:   "admin",
	}

	token, err := mgr.GenerateToken(claims)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	parsedClaims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if parsedClaims.UserID != "123" {
		t.Errorf("expected UserID 123, got %s", parsedClaims.UserID)
	}
	if parsedClaims.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", parsedClaims.Email)
	}
	if parsedClaims.Role != "admin" {
		t.Errorf("expected admin, got %s", parsedClaims.Role)
	}
}

func TestGenerateToken_MissingSecret(t *testing.T) {
	mgr := NewManager("", 1*time.Hour)
	_, err := mgr.GenerateToken(Claims{})
	if !errors.Is(err, ErrMissingSecret) {
		t.Fatalf("expected ErrMissingSecret, got %v", err)
	}
}

func TestValidateToken_MissingSecret(t *testing.T) {
	// First generate a valid token with a secret
	validMgr := NewManager("secret", 1*time.Hour)
	token, _ := validMgr.GenerateToken(Claims{UserID: "1"})

	// Try to validate with a manager that has no secret
	noSecretMgr := NewManager("", 1*time.Hour)
	_, err := noSecretMgr.ValidateToken(token)
	if !errors.Is(err, ErrMissingSecret) {
		t.Fatalf("expected ErrMissingSecret, got %v", err)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	mgr1 := NewManager("secret-one", 1*time.Hour)
	mgr2 := NewManager("secret-two", 1*time.Hour)

	token, _ := mgr1.GenerateToken(Claims{UserID: "1"})

	// Validate with wrong secret
	_, err := mgr2.ValidateToken(token)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// 1 millisecond expiration for fast test
	mgr := NewManager("secret", 1*time.Millisecond)

	token, _ := mgr.GenerateToken(Claims{UserID: "1"})

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	_, err := mgr.ValidateToken(token)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	mgr := NewManager("secret", 1*time.Hour)
	_, err := mgr.ValidateToken("not.a.jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	// Generate a token with a "None" signing method manually to simulate a malicious token
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{UserID: "1"})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	mgr := NewManager("secret", 1*time.Hour)
	_, err := mgr.ValidateToken(tokenString)
	// Our ValidateToken specifically catches non-HMAC methods
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature for none signing method, got %v", err)
	}
}

// ── Package-Level Wrapper Tests ──────────────────────────────────────────────

func TestPackageLevelWrappers(t *testing.T) {
	SetSecret("global-secret")
	SetTokenDuration(1 * time.Hour)

	token, err := GenerateToken(Claims{UserID: "global-user"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if claims.UserID != "global-user" {
		t.Errorf("expected global-user, got %s", claims.UserID)
	}
}

// ── Middleware Tests ─────────────────────────────────────────────────────────

func TestRequireAuth_ValidToken(t *testing.T) {
	mgr := NewManager("test-secret", 1*time.Hour)
	token, _ := mgr.GenerateToken(Claims{UserID: "123"})

	var extractedClaims *Claims
	handler := mgr.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedClaims, _ = ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if extractedClaims == nil {
		t.Fatal("expected claims to be extracted, got nil")
	}
	if extractedClaims.UserID != "123" {
		t.Errorf("expected UserID 123, got %s", extractedClaims.UserID)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	mgr := NewManager("test-secret", 1*time.Hour)

	handler := mgr.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	mgr := NewManager("test-secret", 1*time.Hour)

	handler := mgr.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	// Missing "Bearer "
	req.Header.Set("Authorization", "some-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mgr := NewManager("test-secret", 1*time.Hour)

	handler := mgr.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer bad.token.value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestClaimsFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	claims, ok := ClaimsFromContext(req.Context())
	if ok || claims != nil {
		t.Fatal("expected ok=false and nil claims")
	}
}
