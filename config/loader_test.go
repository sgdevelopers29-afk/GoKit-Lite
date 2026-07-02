package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sgdevelopers29-afk/GoKit-Lite/config"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		fileName    string
		fileContent string
		expectErr   bool
		expectedEnv map[string]string
	}{
		{
			name:        "Successful loading",
			fileName:    "success.env",
			fileContent: "PORT=8080\nAPP_ENV=development\n\n# Database\nDB_HOST=localhost",
			expectErr:   false,
			expectedEnv: map[string]string{
				"PORT":    "8080",
				"APP_ENV": "development",
				"DB_HOST": "localhost",
			},
		},
		{
			name:        "Spaces around values",
			fileName:    "spaces.env",
			fileContent: "PORT = 8080\n APP_ENV = development \n",
			expectErr:   false,
			expectedEnv: map[string]string{
				"PORT":    "8080",
				"APP_ENV": "development",
			},
		},
		{
			name:        "Missing file",
			fileName:    "non_existent.env",
			fileContent: "",
			expectErr:   true,
			expectedEnv: nil,
		},
		{
			name:        "Empty file",
			fileName:    "empty.env",
			fileContent: "",
			expectErr:   false,
			expectedEnv: map[string]string{},
		},
		{
			name:        "Comment lines only",
			fileName:    "comments.env",
			fileContent: "# just a comment\n\n# another comment",
			expectErr:   false,
			expectedEnv: map[string]string{},
		},
		{
			name:        "Invalid format",
			fileName:    "invalid.env",
			fileContent: "PORT=8080\nINVALID_LINE\nAPP_ENV=development",
			expectErr:   true,
			expectedEnv: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tt.fileName)

			// Only create the file if it's not the "Missing file" test case
			if tt.name != "Missing file" {
				err := os.WriteFile(path, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			// Clean up environment variables before each test
			for k := range tt.expectedEnv {
				os.Unsetenv(k)
			}

			err := config.Load(path)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check expected env vars
				for k, v := range tt.expectedEnv {
					actual := os.Getenv(k)
					if actual != v {
						t.Errorf("expected env %s=%s, got %s", k, v, actual)
					}
				}
			}
		})
	}
}
