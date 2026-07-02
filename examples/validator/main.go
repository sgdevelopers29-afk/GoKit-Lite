package main

import (
	"fmt"

	"github.com/sgdevelopers29-afk/GoKit-Lite/validator"
)

type UserRequest struct {
	Username string `required:"true" minLength:"3"`
	Email    string `required:"true" email:"true"`
	Age      int    `required:"true"`
}

func main() {
	// A request with invalid data
	req := UserRequest{
		Username: "ab",            // Fails minLength
		Email:    "invalid-email", // Fails email
		Age:      0,               // Fails required (0 is zero value)
	}

	// ValidateAll collects all errors at once
	result := validator.ValidateAll(req)

	if !result.Valid {
		fmt.Println("Validation failed with the following errors:")
		for _, err := range result.Errors {
			fmt.Printf("- Field '%s': %s\n", err.Field, err.Message)
		}
	} else {
		fmt.Println("Validation passed!")
	}
}
