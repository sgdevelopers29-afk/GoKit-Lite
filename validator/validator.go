// Package validator provides lightweight, reflection-based struct validation
// for Go. It reads validation rules from struct field tags and returns a
// structured [ValidationError] on the first violated rule.
//
// Supported tags (V5):
//
//   - required:"true"       — the field must not be its zero value; for slices
//     and maps this means non-nil and non-empty
//   - min:"<n>"             — numeric field must be >= n
//   - max:"<n>"             — numeric field must be <= n
//   - email:"true"          — string field must be a valid e-mail address
//   - minLength:"<n>"       — string field must have at least n Unicode code points
//   - maxLength:"<n>"       — string field must have at most n Unicode code points
//   - regex:"<pattern>"     — string field must match the given regular expression
//   - oneOf:"<a>,<b>,..."   — string field must be one of the comma-separated values
//   - eqField:"<Field>"     — field value must equal the specified field in the same struct
//
// V5 framework additions:
//
//   - Custom field validators: Register custom rules via [Register].
//   - Cross-field (struct) validators: Register type-level checks via [RegisterStructValidator].
//   - Error aggregation: Use [ValidateAll] to collect all failures via [Result].
//
// Usage:
//
//	type User struct {
//	    Name            string `required:"true"`
//	    Password        string `required:"true" minLength:"8"`
//	    ConfirmPassword string `required:"true" eqField:"Password"`
//	}
//
//	if result := validator.ValidateAll(user); !result.Valid {
//	    for _, e := range result.Errors {
//	        fmt.Println(e.Field, e.Message)
//	    }
//	}
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// maxRecursionDepth is the maximum number of nested struct levels that
// validateStruct will descend into. It guards against pathological inputs
// (e.g. self-referential types via interface fields) without panicking.
const maxRecursionDepth = 32

// emailRegexp is the compiled regular expression used for e-mail validation.
// It follows the most common subset of RFC 5322 that covers real-world addresses.
var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// regexpCache stores previously compiled *regexp.Regexp instances keyed by pattern string.
var regexpCache sync.Map

// ── V5 Public Types and Registration ─────────────────────────────────────────

// ValidatorFunc is the signature for custom field-level validators.
// value is the field's current value. Return true to pass, false to fail.
type ValidatorFunc func(value any) bool

// StructValidatorFunc is the signature for struct-level (cross-field) validators.
// s is the complete struct value. Return true to pass, false to fail.
type StructValidatorFunc func(s any) bool

var (
	// customValidators holds ValidatorFunc keyed by string tag name.
	customValidators sync.Map
	// structValidators holds StructValidatorFunc keyed by reflect.Type.
	structValidators sync.Map
)

// reservedTags are built-in tag names that cannot be overwritten by custom validators.
var reservedTags = map[string]bool{
	"required":  true,
	"min":       true,
	"max":       true,
	"email":     true,
	"minLength": true,
	"maxLength": true,
	"regex":     true,
	"oneOf":     true,
	"eqField":   true,
}

// Register registers a custom field-level validator under the given tag name.
// Returns an error if name is empty, name collides with a built-in tag, or fn is nil.
func Register(name string, fn ValidatorFunc) error {
	if name == "" {
		return errors.New("validator: registration name cannot be empty")
	}
	if reservedTags[name] {
		return fmt.Errorf("validator: cannot overwrite reserved built-in tag %q", name)
	}
	if fn == nil {
		return errors.New("validator: validation function cannot be nil")
	}
	customValidators.Store(name, fn)
	return nil
}

// Unregister removes a previously registered custom field validator.
// No-op if name was never registered.
func Unregister(name string) {
	customValidators.Delete(name)
}

// RegisterStructValidator registers a struct-level cross-field validator for the given
// reflect.Type. Returns an error if t is nil, not a struct, or fn is nil.
func RegisterStructValidator(t reflect.Type, fn StructValidatorFunc) error {
	if t == nil {
		return errors.New("validator: type cannot be nil")
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("validator: struct validator must be registered for a struct type, got %v", t.Kind())
	}
	if fn == nil {
		return errors.New("validator: struct validation function cannot be nil")
	}
	structValidators.Store(t, fn)
	return nil
}

// UnregisterStructValidator removes a previously registered struct validator.
func UnregisterStructValidator(t reflect.Type) {
	structValidators.Delete(t)
}

// ── Validation Entrypoints ────────────────────────────────────────────────────

// Validate inspects every exported field of data using struct tags and returns
// an error for the first rule violation it finds (fail-fast).
//
// data must be a struct or a non-nil pointer to a struct.
func Validate(data any) error {
	v, err := prepareValue(data)
	if err != nil {
		return err
	}

	errs, err := collectErrors(v, "", 0, true)
	if err != nil {
		return err // Programming errors
	}
	if len(errs) > 0 {
		ve := errs[0]
		return &ve
	}
	return nil
}

// ValidateAll inspects every exported field of data using struct tags and returns
// a [*Result] containing all validation errors found.
// Within a single field, only the first failing rule is collected (per-field fail-fast).
// Across fields, all failures are collected.
func ValidateAll(data any) *Result {
	res := &Result{Valid: true}
	v, err := prepareValue(data)
	if err != nil {
		res.Valid = false
		res.Errors = append(res.Errors, ValidationError{
			Field:   "",
			Rule:    "input",
			Message: err.Error(),
		})
		return res
	}

	errs, sysErr := collectErrors(v, "", 0, false)
	if sysErr != nil {
		res.Valid = false
		res.Errors = append(res.Errors, ValidationError{
			Field:   "",
			Rule:    "system",
			Message: sysErr.Error(),
		})
		return res
	}

	if len(errs) > 0 {
		res.Valid = false
		res.Errors = errs
	}

	return res
}

func prepareValue(data any) (reflect.Value, error) {
	if data == nil {
		return reflect.Value{}, errors.New("validator: input cannot be nil")
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, errors.New("validator: input cannot be nil")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("validator: input must be a struct, got %s", v.Kind())
	}

	return v, nil
}

// ── Recursive engine ──────────────────────────────────────────────────────────

func collectErrors(v reflect.Value, prefix string, depth int, failFast bool) ([]ValidationError, error) {
	if depth > maxRecursionDepth {
		return nil, fmt.Errorf("validator: maximum recursion depth (%d) exceeded at %q — possible circular reference",
			maxRecursionDepth, prefix)
	}

	var allErrs []ValidationError
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !field.IsExported() {
			continue
		}

		qualName := qualifiedName(prefix, field.Name)

		// Apply field-level rules
		ve, sysErr := applyFieldRules(field, fieldVal, qualName, v)
		if sysErr != nil {
			return nil, sysErr
		}
		if ve != nil {
			allErrs = append(allErrs, *ve)
			if failFast {
				return allErrs, nil
			}
			// Skip descending into children of this field if the field itself is invalid
			continue
		}

		// Recursive descent
		descVal := fieldVal
		if descVal.Kind() == reflect.Ptr {
			if descVal.IsNil() {
				continue
			}
			descVal = descVal.Elem()
		}

		var nestedErrs []ValidationError
		var err error

		switch descVal.Kind() {
		case reflect.Struct:
			nestedErrs, err = collectErrors(descVal, qualName, depth+1, failFast)
		case reflect.Slice, reflect.Array:
			nestedErrs, err = collectSliceErrors(descVal, qualName, depth, failFast)
		case reflect.Map:
			nestedErrs, err = collectMapErrors(descVal, qualName, depth, failFast)
		}

		if err != nil {
			return nil, err
		}
		if len(nestedErrs) > 0 {
			allErrs = append(allErrs, nestedErrs...)
			if failFast {
				return allErrs, nil
			}
		}
	}

	// Apply struct-level validators for this struct
	if fnv, ok := structValidators.Load(t); ok {
		fn := fnv.(StructValidatorFunc)
		if !fn(v.Interface()) {
			structQualName := prefix
			if structQualName == "" {
				structQualName = t.Name()
			}
			ve := ValidationError{
				Field:   structQualName,
				Rule:    "struct",
				Message: fmt.Sprintf("struct validation failed for %s", structQualName),
			}
			allErrs = append(allErrs, ve)
			if failFast {
				return allErrs, nil
			}
		}
	}

	return allErrs, nil
}

func applyFieldRules(field reflect.StructField, val reflect.Value, qualName string, parentStruct reflect.Value) (*ValidationError, error) {
	if err := applyRequired(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyMin(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyMax(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyEmail(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyMinLength(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyMaxLength(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyRegex(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyOneOf(field, val, qualName); err != nil {
		return extractError(err)
	}
	if err := applyEqField(field, val, qualName, parentStruct); err != nil {
		return extractError(err)
	}
	if err := applyCustomValidators(field, val, qualName); err != nil {
		return extractError(err)
	}
	return nil, nil
}

func extractError(err error) (*ValidationError, error) {
	if err == nil {
		return nil, nil
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve, nil
	}
	return nil, err // System/programming error
}

func collectSliceErrors(v reflect.Value, prefix string, depth int, failFast bool) ([]ValidationError, error) {
	var allErrs []ValidationError
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		elemPrefix := fmt.Sprintf("%s[%d]", prefix, i)

		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		if elem.Kind() == reflect.Struct {
			errs, err := collectErrors(elem, elemPrefix, depth+1, failFast)
			if err != nil {
				return nil, err
			}
			if len(errs) > 0 {
				allErrs = append(allErrs, errs...)
				if failFast {
					return allErrs, nil
				}
			}
		}
	}
	return allErrs, nil
}

func collectMapErrors(v reflect.Value, prefix string, depth int, failFast bool) ([]ValidationError, error) {
	var allErrs []ValidationError
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		valPrefix := fmt.Sprintf("%s[%v]", prefix, key)

		if val.Kind() == reflect.Interface {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}

		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			errs, err := collectErrors(val, valPrefix, depth+1, failFast)
			if err != nil {
				return nil, err
			}
			if len(errs) > 0 {
				allErrs = append(allErrs, errs...)
				if failFast {
					return allErrs, nil
				}
			}
		}
	}
	return allErrs, nil
}

func qualifiedName(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// ── Rule implementations ──────────────────────────────────────────────────────

func applyRequired(field reflect.StructField, val reflect.Value, qualName string) error {
	if field.Tag.Get("required") != "true" {
		return nil
	}
	if isZeroValue(val) {
		return newValidationError(
			qualName,
			"required",
			fmt.Sprintf("field %s is required but has a zero value", qualName),
		)
	}
	return nil
}

func applyMin(field reflect.StructField, val reflect.Value, qualName string) error {
	tag := field.Tag.Get("min")
	if tag == "" {
		return nil
	}

	limit, err := strconv.Atoi(tag)
	if err != nil {
		return fmt.Errorf(
			"validator: field %s has an invalid min tag value %q (must be an integer): %w",
			qualName, tag, err,
		)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() < int64(limit) {
			return newValidationError(
				qualName,
				"min",
				fmt.Sprintf("field %s must be >= %d, got %d", qualName, limit, val.Int()),
			)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if limit < 0 {
			return nil
		}
		if val.Uint() < uint64(limit) {
			return newValidationError(
				qualName,
				"min",
				fmt.Sprintf("field %s must be >= %d, got %d", qualName, limit, val.Uint()),
			)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() < float64(limit) {
			return newValidationError(
				qualName,
				"min",
				fmt.Sprintf("field %s must be >= %d, got %g", qualName, limit, val.Float()),
			)
		}
	}
	return nil
}

func applyMax(field reflect.StructField, val reflect.Value, qualName string) error {
	tag := field.Tag.Get("max")
	if tag == "" {
		return nil
	}

	limit, err := strconv.Atoi(tag)
	if err != nil {
		return fmt.Errorf(
			"validator: field %s has an invalid max tag value %q (must be an integer): %w",
			qualName, tag, err,
		)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() > int64(limit) {
			return newValidationError(
				qualName,
				"max",
				fmt.Sprintf("field %s must be <= %d, got %d", qualName, limit, val.Int()),
			)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if limit < 0 {
			if val.Uint() > 0 || uint64(0) > uint64(limit) {
				return newValidationError(
					qualName,
					"max",
					fmt.Sprintf("field %s must be <= %d, got %d", qualName, limit, val.Uint()),
				)
			}
		} else if val.Uint() > uint64(limit) {
			return newValidationError(
				qualName,
				"max",
				fmt.Sprintf("field %s must be <= %d, got %d", qualName, limit, val.Uint()),
			)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() > float64(limit) {
			return newValidationError(
				qualName,
				"max",
				fmt.Sprintf("field %s must be <= %d, got %g", qualName, limit, val.Float()),
			)
		}
	}
	return nil
}

func applyEmail(field reflect.StructField, val reflect.Value, qualName string) error {
	if field.Tag.Get("email") != "true" {
		return nil
	}
	if val.Kind() != reflect.String {
		return fmt.Errorf(
			"validator: email tag is only valid on string fields, field %s is %s",
			qualName, val.Kind(),
		)
	}
	if !emailRegexp.MatchString(val.String()) {
		return newValidationError(
			qualName,
			"email",
			fmt.Sprintf("field %s must be a valid email address", qualName),
		)
	}
	return nil
}

func applyMinLength(field reflect.StructField, val reflect.Value, qualName string) error {
	tag := field.Tag.Get("minLength")
	if tag == "" {
		return nil
	}

	limit, err := strconv.Atoi(tag)
	if err != nil {
		return fmt.Errorf(
			"validator: field %s has an invalid minLength tag value %q (must be a non-negative integer): %w",
			qualName, tag, err,
		)
	}
	if limit < 0 {
		return fmt.Errorf(
			"validator: field %s has a negative minLength tag value %d",
			qualName, limit,
		)
	}

	if val.Kind() != reflect.String {
		return nil
	}

	if utf8.RuneCountInString(val.String()) < limit {
		return newValidationError(
			qualName,
			"minLength",
			fmt.Sprintf("field %s must have at least %d character(s), got %d",
				qualName, limit, utf8.RuneCountInString(val.String())),
		)
	}
	return nil
}

func applyMaxLength(field reflect.StructField, val reflect.Value, qualName string) error {
	tag := field.Tag.Get("maxLength")
	if tag == "" {
		return nil
	}

	limit, err := strconv.Atoi(tag)
	if err != nil {
		return fmt.Errorf(
			"validator: field %s has an invalid maxLength tag value %q (must be a non-negative integer): %w",
			qualName, tag, err,
		)
	}
	if limit < 0 {
		return fmt.Errorf(
			"validator: field %s has a negative maxLength tag value %d",
			qualName, limit,
		)
	}

	if val.Kind() != reflect.String {
		return nil
	}

	if utf8.RuneCountInString(val.String()) > limit {
		return newValidationError(
			qualName,
			"maxLength",
			fmt.Sprintf("field %s must have at most %d character(s), got %d",
				qualName, limit, utf8.RuneCountInString(val.String())),
		)
	}
	return nil
}

func applyRegex(field reflect.StructField, val reflect.Value, qualName string) error {
	pattern := field.Tag.Get("regex")
	if pattern == "" {
		return nil
	}

	if val.Kind() != reflect.String {
		return nil
	}

	var re *regexp.Regexp
	if cached, ok := regexpCache.Load(pattern); ok {
		re = cached.(*regexp.Regexp)
	} else {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf(
				"validator: field %s has an invalid regex tag pattern %q: %w",
				qualName, pattern, err,
			)
		}
		regexpCache.Store(pattern, compiled)
		re = compiled
	}

	if !re.MatchString(val.String()) {
		return newValidationError(
			qualName,
			"regex",
			fmt.Sprintf("field %s must match pattern %q", qualName, pattern),
		)
	}
	return nil
}

func applyOneOf(field reflect.StructField, val reflect.Value, qualName string) error {
	tag := field.Tag.Get("oneOf")
	if tag == "" {
		return nil
	}

	if val.Kind() != reflect.String {
		return nil
	}

	allowed := strings.Split(tag, ",")
	got := val.String()

	for _, choice := range allowed {
		if strings.TrimSpace(choice) == got {
			return nil
		}
	}

	return newValidationError(
		qualName,
		"oneOf",
		fmt.Sprintf("field %s must be one of [%s], got %q", qualName, tag, got),
	)
}

func applyEqField(field reflect.StructField, val reflect.Value, qualName string, parentStruct reflect.Value) error {
	targetFieldName := field.Tag.Get("eqField")
	if targetFieldName == "" {
		return nil
	}

	targetField := parentStruct.FieldByName(targetFieldName)
	if !targetField.IsValid() {
		return fmt.Errorf("validator: eqField tag on %s points to non-existent field %q", qualName, targetFieldName)
	}

	if !reflect.DeepEqual(val.Interface(), targetField.Interface()) {
		return newValidationError(
			qualName,
			"eqField",
			fmt.Sprintf("field %s must be equal to field %s", qualName, targetFieldName),
		)
	}

	return nil
}

func applyCustomValidators(field reflect.StructField, val reflect.Value, qualName string) error {
	var firstErr error
	customValidators.Range(func(key, value any) bool {
		tagName := key.(string)
		tagVal := field.Tag.Get(tagName)
		if tagVal == "true" { // only activate if value is "true"
			fn := value.(ValidatorFunc)
			if !fn(val.Interface()) {
				firstErr = newValidationError(
					qualName,
					tagName,
					fmt.Sprintf("field %s failed custom validation %q", qualName, tagName),
				)
				return false // stop range iteration and return this error
			}
		}
		return true
	})
	return firstErr
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// isZeroValue reports whether v holds the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}
