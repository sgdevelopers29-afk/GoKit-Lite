package validator

import (
	"errors"
	"fmt"
	"reflect"
)

// Validate validates a struct based on validation tags.
func Validate(data any) error {

	if data == nil {
		return errors.New("input cannot be nil")
	}

	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	// Support pointers
	if t.Kind() == reflect.Ptr {

		if v.IsNil() {
			return errors.New("input cannot be nil")
		}

		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}

	for i := 0; i < t.NumField(); i++ {

		field := t.Field(i)
		fieldValue := v.Field(i)

		required := field.Tag.Get("required")

		if required == "true" {

			if isZeroValue(fieldValue) {
				return fmt.Errorf(
					"field %s is required",
					field.Name,
				)
			}
		}
	}

	return nil
}

// isZeroValue checks whether a field contains its zero value.
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
