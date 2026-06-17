package validator_test

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sgdevelopers29-afk/GoKit-Lite/validator"
)

func ExampleValidate() {
	type User struct {
		Name  string `required:"true"`
		Email string `required:"true" email:"true"`
		Age   int    `min:"18"`
	}

	// This user is invalid: Name is empty, Email is invalid, Age < 18.
	// Validate operates in fail-fast mode, so it returns the first error found.
	u := User{
		Name:  "",
		Email: "invalid-email",
		Age:   16,
	}

	err := validator.Validate(u)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: field Name failed rule "required": field Name is required but has a zero value
}

func ExampleValidateAll() {
	type User struct {
		Name  string `required:"true"`
		Email string `required:"true" email:"true"`
		Age   int    `min:"18"`
	}

	// ValidateAll collects all errors across different fields.
	// Within a single field, it still stops at the first failing rule.
	u := User{
		Name:  "",
		Email: "invalid-email",
		Age:   16,
	}

	result := validator.ValidateAll(u)
	if !result.Valid {
		fmt.Printf("Found %d errors:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("- %s: %s\n", err.Field, err.Rule)
		}
	}
	// Output:
	// Found 3 errors:
	// - Name: required
	// - Email: email
	// - Age: min
}

func ExampleRegister() {
	// Register a custom field-level validator
	validator.Register("username", func(value any) bool {
		s, ok := value.(string)
		if !ok {
			return false
		}
		// Custom logic: username must start with "user_"
		return strings.HasPrefix(s, "user_")
	})
	defer validator.Unregister("username")

	type Account struct {
		Username string `required:"true" username:"true"`
	}

	acc := Account{Username: "admin"}
	err := validator.Validate(acc)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: field Username failed rule "username": field Username failed custom validation "username"
}

func ExampleRegisterStructValidator() {
	type DateRange struct {
		Start int
		End   int
	}

	// Register a struct-level cross-field validator keyed by reflect.Type
	t := reflect.TypeOf(DateRange{})
	validator.RegisterStructValidator(t, func(s any) bool {
		dr, ok := s.(DateRange)
		if !ok {
			return false
		}
		// Custom cross-field logic: End must be >= Start
		return dr.End >= dr.Start
	})
	defer validator.UnregisterStructValidator(t)

	dr := DateRange{Start: 2024, End: 2023}
	err := validator.Validate(dr)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: field DateRange failed rule "struct": struct validation failed for DateRange
}

func Example_eqField() {
	type ResetPassword struct {
		NewPassword     string `required:"true" minLength:"8"`
		ConfirmPassword string `required:"true" eqField:"NewPassword"`
	}

	req := ResetPassword{
		NewPassword:     "secure-password",
		ConfirmPassword: "different-password",
	}

	err := validator.Validate(req)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output:
	// Error: field ConfirmPassword failed rule "eqField": field ConfirmPassword must be equal to field NewPassword
}
