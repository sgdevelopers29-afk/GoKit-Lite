package validator

import "fmt"

// ValidationError represents a single field-level validation failure.
//
// Field holds the struct field name that failed validation.
// Rule holds the name of the rule that was violated (e.g. "required", "min", "max", "email").
//
// Callers can use errors.As to extract a *ValidationError from the error
// returned by Validate and inspect the individual Field and Rule values.
//
// Example:
//
//	var ve *ValidationError
//	if errors.As(err, &ve) {
//	    fmt.Println(ve.Field, ve.Rule)
//	}
type ValidationError struct {
	// Field is the name of the struct field that failed validation.
	Field string

	// Rule is the name of the validation rule that was violated.
	Rule string

	// Message is a human-readable description of the failure.
	Message string
}

// Error implements the error interface.
// The returned string is in the form: "field <Field> failed rule '<Rule>': <Message>".
func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %s failed rule %q: %s", e.Field, e.Rule, e.Message)
}

// newValidationError constructs a *ValidationError and returns it as an error.
// This helper keeps rule implementations concise.
func newValidationError(field, rule, message string) error {
	return &ValidationError{
		Field:   field,
		Rule:    rule,
		Message: message,
	}
}
