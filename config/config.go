// Package config provides configuration management for the GoKit-Lite toolkit.
package config

import "os"

// Get retrieves the value of the environment variable named by the key.
// It returns an empty string if the variable is not present.
// This function wraps os.Getenv to provide a centralized access point
// for future configuration enhancements.
func Get(key string) string {
	return os.Getenv(key)
}

// GetOrDefault retrieves the value of the environment variable named by the key.
// If the variable is not present or is empty, it returns the provided defaultVal.
func GetOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
