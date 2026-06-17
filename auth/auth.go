package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultDuration = 24 * time.Hour

// defaultManager is the internal instance used by package-level wrapper functions.
var defaultManager = NewManager("", defaultDuration)

// Manager handles the creation and validation of JWT tokens.
// It is safe for concurrent use.
type Manager struct {
	secret        []byte
	tokenDuration time.Duration
}

// NewManager creates a new JWT Manager with a specific secret and token duration.
// Providing an empty secret is allowed at creation time, but attempting to generate
// or validate tokens without a secret will return ErrMissingSecret.
func NewManager(secret string, duration time.Duration) *Manager {
	if duration <= 0 {
		duration = defaultDuration
	}
	return &Manager{
		secret:        []byte(secret),
		tokenDuration: duration,
	}
}

// SetSecret sets the secret key used for signing and validating tokens on the Manager.
func (m *Manager) SetSecret(secret string) {
	m.secret = []byte(secret)
}

// SetTokenDuration sets the default expiration duration for new tokens on the Manager.
func (m *Manager) SetTokenDuration(d time.Duration) {
	if d > 0 {
		m.tokenDuration = d
	}
}

// GenerateToken creates a signed JWT string from the provided Claims.
// It automatically populates the IssuedAt (iat) and ExpiresAt (exp) fields
// if they are not already set.
func (m *Manager) GenerateToken(claims Claims) (string, error) {
	if len(m.secret) == 0 {
		return "", ErrMissingSecret
	}

	now := time.Now()

	// Only set IAT if not provided
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(now)
	}

	// Only set EXP if not provided
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(m.tokenDuration))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

// ValidateToken parses a JWT string, verifies its signature using the configured secret,
// ensures it has not expired, and returns the parsed Claims.
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	if len(m.secret) == 0 {
		return nil, ErrMissingSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect (HMAC).
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return m.secret, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenSignatureInvalid) || errors.Is(err, ErrInvalidSignature):
			return nil, ErrInvalidSignature
		default:
			return nil, ErrInvalidToken
		}
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ── Package-Level Wrappers ───────────────────────────────────────────────────

// SetSecret sets the secret key used by the default package-level auth manager.
func SetSecret(secret string) {
	defaultManager.SetSecret(secret)
}

// SetTokenDuration sets the default expiration duration for new tokens on the default manager.
// If not explicitly set, the duration defaults to 24 hours.
func SetTokenDuration(d time.Duration) {
	defaultManager.SetTokenDuration(d)
}

// GenerateToken creates a signed JWT string using the default package-level manager.
func GenerateToken(claims Claims) (string, error) {
	return defaultManager.GenerateToken(claims)
}

// ValidateToken parses and verifies a JWT string using the default package-level manager.
func ValidateToken(tokenString string) (*Claims, error) {
	return defaultManager.ValidateToken(tokenString)
}
