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

// Result holds the aggregated outcome of a ValidateAll call.
type Result struct {
	// Valid is true when no validation rules were violated.
	Valid bool

	// Errors contains every ValidationError collected during validation.
	// Empty when Valid is true.
	Errors []ValidationError
}

// Error implements the error interface so *Result can be used wherever
// an error is expected. If the result is valid, Error returns a generic string,
// but usually callers check Valid or compare against nil.
// If invalid, it returns a combined string of all errors.
func (r *Result) Error() string {
	if r.Valid || len(r.Errors) == 0 {
		return "validator: valid"
	}
	var msgs []string
	for _, e := range r.Errors {
		msgs = append(msgs, e.Error())
	}
	return fmt.Sprintf("%d validation errors: %v", len(r.Errors), msgs)
}

// First returns the first ValidationError as a *ValidationError pointer,
// or nil if there are no errors. This is useful for fail-fast callers upgrading
// to ValidateAll.
func (r *Result) First() *ValidationError {
	if r.Valid || len(r.Errors) == 0 {
		return nil
	}
	return &r.Errors[0]
}
