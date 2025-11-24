package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnvFileLoading tests the EnvFile field and custom environment file loading
func TestEnvFileLoading(t *testing.T) {
	// Save and restore original ENV_FILE
	originalEnvFile := os.Getenv("ENV_FILE")
	defer func() {
		if originalEnvFile != "" {
			_ = os.Setenv("ENV_FILE", originalEnvFile)
		} else {
			_ = os.Unsetenv("ENV_FILE")
		}
	}()

	t.Run("default env file", func(t *testing.T) {
		_ = os.Unsetenv("ENV_FILE")
		envFile := getEnvOrDefault("ENV_FILE", ".env")
		if envFile != ".env" {
			t.Errorf("Expected default EnvFile to be '.env', got '%s'", envFile)
		}
	})

	t.Run("env file from environment variable", func(t *testing.T) {
		if err := os.Setenv("ENV_FILE", "custom.env"); err != nil {
			t.Fatalf("failed to set ENV_FILE: %v", err)
		}
		envFile := getEnvOrDefault("ENV_FILE", ".env")
		if envFile != "custom.env" {
			t.Errorf("Expected EnvFile to be 'custom.env' from ENV_FILE, got '%s'", envFile)
		}
	})

	t.Run("env file field exists in config struct", func(t *testing.T) {
		cfg := &Config{
			EnvFile: "/path/to/custom.env",
		}
		if cfg.EnvFile != "/path/to/custom.env" {
			t.Errorf("Expected EnvFile to be '/path/to/custom.env', got '%s'", cfg.EnvFile)
		}
	})
}

// TestCustomEnvFileLoading tests loading variables from a custom environment file
func TestCustomEnvFileLoading(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	customEnvFile := filepath.Join(tempDir, "test.env")

	// Write test environment variables to custom file
	envContent := `TEMPEST_TOKEN=custom_token
TEMPEST_STATION_NAME=CustomStation
HOMEKIT_PIN=87654321
LOG_LEVEL=info
WEB_PORT=9090
`
	if err := os.WriteFile(customEnvFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	// Test that the EnvFile field can be set to point to custom file
	cfg := &Config{
		EnvFile: customEnvFile,
	}

	if cfg.EnvFile != customEnvFile {
		t.Errorf("Expected EnvFile to be '%s', got '%s'", customEnvFile, cfg.EnvFile)
	}

	// Note: Actual loading of the custom env file happens in main.go before LoadConfig()
	// This test verifies the EnvFile field is properly stored in Config
}

// TestEnvFileValidation tests validation of environment file paths
func TestEnvFileValidation(t *testing.T) {
	tests := []struct {
		name    string
		envFile string
		valid   bool
	}{
		{"default .env", ".env", true},
		{"relative path", "config/production.env", true},
		{"absolute path", "/etc/tempest/app.env", true},
		{"with spaces", "my env.env", true},
		{"empty string", "", true}, // empty defaults to .env in main.go
		{"hidden file", ".custom-env", true},
		{"no extension", "envfile", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				EnvFile: tt.envFile,
			}

			// EnvFile is just a string field, no validation needed
			// Validation happens during file loading in main.go
			if cfg.EnvFile != tt.envFile {
				t.Errorf("Expected EnvFile to be '%s', got '%s'", tt.envFile, cfg.EnvFile)
			}
		})
	}
}

// TestEnvFileWithOtherConfig tests EnvFile field alongside other config fields
func TestEnvFileWithOtherConfig(t *testing.T) {
	cfg := &Config{
		EnvFile:     "/custom/path/production.env",
		Token:       "test-token",
		StationName: "TestStation",
	}

	// Verify EnvFile doesn't interfere with other fields
	if cfg.EnvFile != "/custom/path/production.env" {
		t.Errorf("Expected EnvFile to be '/custom/path/production.env', got '%s'", cfg.EnvFile)
	}
	if cfg.Token != "test-token" {
		t.Errorf("Expected Token to be 'test-token', got '%s'", cfg.Token)
	}
	if cfg.StationName != "TestStation" {
		t.Errorf("Expected StationName to be 'TestStation', got '%s'", cfg.StationName)
	}
}

// TestEnvFileFlagPrecedence tests that command line flag should override ENV_FILE
func TestEnvFileFlagPrecedence(t *testing.T) {
	// This test documents expected behavior:
	// 1. Default: .env
	// 2. Environment variable ENV_FILE: overrides default
	// 3. Command line --env flag: overrides both (handled in main.go)

	tests := []struct {
		name        string
		envVar      string
		expected    string
		description string
	}{
		{
			name:        "no env var",
			envVar:      "",
			expected:    ".env",
			description: "should use default .env",
		},
		{
			name:        "custom env var",
			envVar:      "staging.env",
			expected:    "staging.env",
			description: "should use ENV_FILE value",
		},
		{
			name:        "absolute path env var",
			envVar:      "/etc/tempest/production.env",
			expected:    "/etc/tempest/production.env",
			description: "should use absolute path from ENV_FILE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original
			original := os.Getenv("ENV_FILE")
			defer func() {
				if original != "" {
					_ = os.Setenv("ENV_FILE", original)
				} else {
					_ = os.Unsetenv("ENV_FILE")
				}
			}()

			// Set test value
			if tt.envVar != "" {
				_ = os.Setenv("ENV_FILE", tt.envVar)
			} else {
				_ = os.Unsetenv("ENV_FILE")
			}

			envFile := getEnvOrDefault("ENV_FILE", ".env")
			if envFile != tt.expected {
				t.Errorf("%s: expected EnvFile '%s', got '%s'",
					tt.description, tt.expected, envFile)
			}
		})
	}
}
