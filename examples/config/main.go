package main

import (
	"fmt"
	"os"

	"github.com/sgdevelopers29-afk/GoKit-Lite/config"
)

func main() {
	// Create a dummy .env file for the example
	os.WriteFile(".env", []byte("PORT=9090\nDEBUG=true"), 0644)
	defer os.Remove(".env")

	// 1. Load the .env file if it exists
	if err := config.LoadIfExists(".env"); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// 2. Retrieve a value
	port := config.Get("PORT")
	fmt.Printf("PORT is: %s\n", port)

	// 3. Retrieve a value with a fallback default
	host := config.GetOrDefault("HOST", "localhost")
	fmt.Printf("HOST is: %s (using default)\n", host)
}
