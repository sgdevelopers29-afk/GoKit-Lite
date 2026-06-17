package auth

import "errors"

var (
	// ErrInvalidToken indicates that the provided JWT is malformed or fundamentally invalid.
	ErrInvalidToken = errors.New("auth: invalid token")

	// ErrExpiredToken indicates that the JWT has expired (the 'exp' claim is in the past).
	ErrExpiredToken = errors.New("auth: token has expired")

	// ErrInvalidSignature indicates that the JWT's signature could not be verified.
	ErrInvalidSignature = errors.New("auth: invalid signature")

	// ErrMissingSecret indicates that no secret key was provided to the Manager for generation or validation.
	ErrMissingSecret = errors.New("auth: missing secret key")

	// ErrMissingAuthHeader indicates that the Authorization header is missing in an HTTP request.
	ErrMissingAuthHeader = errors.New("auth: missing Authorization header")

	// ErrInvalidAuthHeader indicates that the Authorization header does not use the expected 'Bearer <token>' format.
	ErrInvalidAuthHeader = errors.New("auth: invalid Authorization header format")

	// ErrInvalidClaims indicates that the token claims could not be parsed into the required structure.
	ErrInvalidClaims = errors.New("auth: invalid token claims")
)
