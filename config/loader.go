package config

import (
	"fmt"
	"os"
	"strings"
)

// Load reads an environment file from the given path and loads its contents into the system environment.
// It parses KEY=VALUE pairs, ignoring empty lines and comments starting with '#'.
// It safely trims whitespace around keys and values.
// Returns an error if the file cannot be read or if a line contains an invalid format.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Ignore empty lines and comments
		if len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Split on the first '=' character
		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format on line %d: %s", i+1, trimmedLine)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	return nil
}
