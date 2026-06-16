// Package validator provides lightweight, reflection-based struct validation
// for Go. It reads validation rules from struct field tags and returns a
// structured [ValidationError] on the first violated rule.
//
// Supported tags (V2):
//
//   - required:"true"  — the field must not be its zero value
//   - min:"<n>"        — numeric field must be >= n
//   - max:"<n>"        — numeric field must be <= n
//   - email:"true"     — string field must be a valid e-mail address
//
// Usage:
//
//	type User struct {
//	    Name  string `required:"true"`
//	    Age   int    `min:"18" max:"120"`
//	    Email string `required:"true" email:"true"`
//	}
//
//	if err := validator.Validate(user); err != nil {
//	    var ve *validator.ValidationError
//	    if errors.As(err, &ve) {
//	        fmt.Println(ve.Field, ve.Rule)
//	    }
//	}
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// emailRegexp is the compiled regular expression used for e-mail validation.
// It follows the most common subset of RFC 5322 that covers real-world addresses.
// Compiled once at package initialisation to avoid per-call overhead.
var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Validate inspects every exported field of data using struct tags and returns
// a [*ValidationError] for the first rule violation it finds.
//
// data must be a struct or a non-nil pointer to a struct; any other value
// (including nil) causes an error to be returned immediately.
//
// Validation rules are applied in tag-declaration order:
//  1. required
//  2. min
//  3. max
//  4. email
//
// Validate stops and returns on the very first failure (fail-fast). If all
// fields pass, nil is returned.
func Validate(data any) error {

	if data == nil {
		return errors.New("validator: input cannot be nil")
	}

	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	// Dereference a pointer exactly one level so callers can pass either T or *T.
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return errors.New("validator: input cannot be nil")
		}
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("validator: input must be a struct, got %s", t.Kind())
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Skip unexported fields — reflect cannot reliably read their values
		// and tag-based validation on unexported fields is a misuse of the API.
		if !field.IsExported() {
			continue
		}

		if err := applyRequired(field, fieldVal); err != nil {
			return err
		}

		if err := applyMin(field, fieldVal); err != nil {
			return err
		}

		if err := applyMax(field, fieldVal); err != nil {
			return err
		}

		if err := applyEmail(field, fieldVal); err != nil {
			return err
		}
	}

	return nil
}

// ── Rule implementations ──────────────────────────────────────────────────────

// applyRequired returns a ValidationError when the field carries
// `required:"true"` and its current value is the zero value for its type.
func applyRequired(field reflect.StructField, val reflect.Value) error {
	if field.Tag.Get("required") != "true" {
		return nil
	}
	if isZeroValue(val) {
		return newValidationError(
			field.Name,
			"required",
			fmt.Sprintf("field %s is required but has a zero value", field.Name),
		)
	}
	return nil
}

// applyMin returns a ValidationError when the field carries a `min:"<n>"` tag
// and the field's numeric value is strictly less than n.
//
// The tag is only evaluated on signed integer, unsigned integer, and
// floating-point kinds. For any other kind the tag is silently ignored so that
// adding min to a non-numeric field does not panic.
func applyMin(field reflect.StructField, val reflect.Value) error {
	tag := field.Tag.Get("min")
	if tag == "" {
		return nil
	}

	limit, err := strconv.Atoi(tag)
	if err != nil {
		// Malformed tag — surface a clear error rather than silently ignoring it.
		return fmt.Errorf(
			"validator: field %s has an invalid min tag value %q (must be an integer): %w",
			field.Name, tag, err,
		)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() < int64(limit) {
			return newValidationError(
				field.Name,
				"min",
				fmt.Sprintf("field %s must be >= %d, got %d", field.Name, limit, val.Int()),
			)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if limit < 0 {
			// A negative min on an unsigned field can never be violated.
			return nil
		}
		if val.Uint() < uint64(limit) {
			return newValidationError(
				field.Name,
				"min",
				fmt.Sprintf("field %s must be >= %d, got %d", field.Name, limit, val.Uint()),
			)
		}

	case reflect.Float32, reflect.Float64:
		if val.Float() < float64(limit) {
			return newValidationError(
				field.Name,
				"min",
				fmt.Sprintf("field %s must be >= %d, got %g", field.Name, limit, val.Float()),
			)
		}

	// For all other kinds (string, bool, slice, …) the min tag is a no-op.
	}

	return nil
}

// applyMax returns a ValidationError when the field carries a `max:"<n>"` tag
// and the field's numeric value is strictly greater than n.
//
// The same kind-guarding logic as applyMin applies.
func applyMax(field reflect.StructField, val reflect.Value) error {
	tag := field.Tag.Get("max")
	if tag == "" {
		return nil
	}

	limit, err := strconv.Atoi(tag)
	if err != nil {
		return fmt.Errorf(
			"validator: field %s has an invalid max tag value %q (must be an integer): %w",
			field.Name, tag, err,
		)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() > int64(limit) {
			return newValidationError(
				field.Name,
				"max",
				fmt.Sprintf("field %s must be <= %d, got %d", field.Name, limit, val.Int()),
			)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if limit < 0 {
			// Any unsigned value violates a negative max.
			if val.Uint() > 0 || uint64(0) > uint64(limit) {
				return newValidationError(
					field.Name,
					"max",
					fmt.Sprintf("field %s must be <= %d, got %d", field.Name, limit, val.Uint()),
				)
			}
		} else if val.Uint() > uint64(limit) {
			return newValidationError(
				field.Name,
				"max",
				fmt.Sprintf("field %s must be <= %d, got %d", field.Name, limit, val.Uint()),
			)
		}

	case reflect.Float32, reflect.Float64:
		if val.Float() > float64(limit) {
			return newValidationError(
				field.Name,
				"max",
				fmt.Sprintf("field %s must be <= %d, got %g", field.Name, limit, val.Float()),
			)
		}
	}

	return nil
}

// applyEmail returns a ValidationError when the field carries `email:"true"`
// and the field's string value does not match a valid e-mail format.
//
// The tag is only evaluated on string kinds. For any other kind the tag is
// silently ignored.
func applyEmail(field reflect.StructField, val reflect.Value) error {
	if field.Tag.Get("email") != "true" {
		return nil
	}
	if val.Kind() != reflect.String {
		// email:"true" on a non-string field is a programming error; surface it.
		return fmt.Errorf(
			"validator: email tag is only valid on string fields, field %s is %s",
			field.Name, val.Kind(),
		)
	}
	if !emailRegexp.MatchString(val.String()) {
		return newValidationError(
			field.Name,
			"email",
			fmt.Sprintf("field %s must be a valid email address", field.Name),
		)
	}
	return nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// isZeroValue reports whether v holds the zero value for its type.
// This is used by the required rule.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {

	case reflect.String:
		return v.String() == ""

	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return v.Int() == 0

	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return v.Uint() == 0

	case reflect.Bool:
		return !v.Bool()

	case reflect.Float32,
		reflect.Float64:
		return v.Float() == 0

	case reflect.Slice,
		reflect.Array,
		reflect.Map:
		return v.Len() == 0

	case reflect.Ptr,
		reflect.Interface:
		return v.IsNil()

	default:
		return v.IsZero()
	}
}
