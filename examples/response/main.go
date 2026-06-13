package main

import (
	"encoding/json"
	"fmt"

	"github.com/sgdevelopers29-afk/GoKit-Lite/response"
)

// This example demonstrates how to use the response package to create a standard API response format.
func main() {

	user := map[string]string{
		"name": "Ganesh",
	}

	resp := response.Success(user)

	b, _ := json.MarshalIndent(resp, "", "  ")

	fmt.Println(string(b))
}
