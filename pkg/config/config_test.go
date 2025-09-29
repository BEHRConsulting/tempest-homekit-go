package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateConfigDefaults ensures that validateConfig applies default unit values
// when cfg.Units or cfg.UnitsPressure are empty. This prevents regressions when
// Config structs are constructed programmatically (e.g., in tests).
func TestValidateConfigDefaults(t *testing.T) {
	cfg := &Config{
		Token:         "dummy-token",
		StationName:   "TestStation",
		Pin:           "12345678",
		LogLevel:      "error",
		WebPort:       "8080",
		Sensors:       "temp",
		Units:         "",
		UnitsPressure: "",
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("expected validateConfig to succeed, got error: %v", err)
	}

	if cfg.Units != "imperial" {
		t.Fatalf("expected Units to default to 'imperial', got '%s'", cfg.Units)
	}

	if cfg.UnitsPressure != "inHg" {
		t.Fatalf("expected UnitsPressure to default to 'inHg', got '%s'", cfg.UnitsPressure)
	}
}

func TestLoadConfig(t *testing.T) {
	// Set env vars
	os.Setenv("TEMPEST_TOKEN", "testtoken")
	os.Setenv("TEMPEST_STATION_NAME", "teststation")
	os.Setenv("HOMEKIT_PIN", "12345678")
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("TEMPEST_TOKEN")
	defer os.Unsetenv("TEMPEST_STATION_NAME")
	defer os.Unsetenv("HOMEKIT_PIN")
	defer os.Unsetenv("LOG_LEVEL")

	cfg := LoadConfig()
	if cfg.Token != "testtoken" {
		t.Errorf("Expected token testtoken, got %s", cfg.Token)
	}
	if cfg.StationName != "teststation" {
		t.Errorf("Expected station teststation, got %s", cfg.StationName)
	}
	if cfg.Pin != "12345678" {
		t.Errorf("Expected pin 12345678, got %s", cfg.Pin)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected loglevel debug, got %s", cfg.LogLevel)
	}
}

func TestParseSensorConfigAll(t *testing.T) {
	config := ParseSensorConfig("all")
	expected := SensorConfig{
		Temperature: true,
		Humidity:    true,
		Light:       true,
		Wind:        true,
		Rain:        true,
		Pressure:    true,
		UV:          true,
		Lightning:   true,
	}
	if config != expected {
		t.Errorf("Expected all sensors enabled, got %+v", config)
	}
}

func TestParseSensorConfigMin(t *testing.T) {
	config := ParseSensorConfig("min")
	expected := SensorConfig{
		Temperature: true,
		Humidity:    true,
		Light:       true,
		Wind:        false,
		Rain:        false,
		Pressure:    false,
		UV:          false,
		Lightning:   false,
	}
	if config != expected {
		t.Errorf("Expected min sensor config, got %+v", config)
	}
}

func TestParseSensorConfigTempOnly(t *testing.T) {
	config := ParseSensorConfig("temp")
	expected := SensorConfig{
		Temperature: true,
		Humidity:    false,
		Light:       false,
		Wind:        false,
		Rain:        false,
		Pressure:    false,
		UV:          false,
		Lightning:   false,
	}
	if config != expected {
		t.Errorf("Expected temp-only config, got %+v", config)
	}
}

func TestParseSensorConfigCustom(t *testing.T) {
	config := ParseSensorConfig("temp,humidity,wind")
	expected := SensorConfig{
		Temperature: true,
		Humidity:    true,
		Light:       false,
		Wind:        true,
		Rain:        false,
		Pressure:    false,
		UV:          false,
		Lightning:   false,
	}
	if config != expected {
		t.Errorf("Expected custom sensor config, got %+v", config)
	}
}

func TestParseElevationFeet(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"903ft", 275.2344, false}, // 903 feet = 275.2344 meters (more precise)
		{"1000ft", 304.8, false},   // 1000 feet = 304.8 meters
		{"0ft", 0.0, false},
		{"100.5ft", 30.6324, false}, // 100.5 feet = 30.6324 meters
	}

	for _, test := range tests {
		result, err := parseElevation(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
			// Allow small floating point differences
			if abs(result-test.expected) > 0.01 {
				t.Errorf("For input %s, expected %.4f meters, got %.4f", test.input, test.expected, result)
			}
		}
	}
}

func TestParseElevationMeters(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"275m", 275.0, false},
		{"100.5m", 100.5, false},
		{"0m", 0.0, false},
		{"1000m", 1000.0, false},
	}

	for _, test := range tests {
		result, err := parseElevation(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
			if abs(result-test.expected) > 0.01 {
				t.Errorf("For input %s, expected %.4f meters, got %.4f", test.input, test.expected, result)
			}
		}
	}
}

func TestParseElevationNoUnit(t *testing.T) {
	// Should assume meters when no unit provided
	result, err := parseElevation("275")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != 275.0 {
		t.Errorf("Expected 275.0 meters, got %.4f", result)
	}
}

func TestParseElevationInvalid(t *testing.T) {
	invalidInputs := []string{
		"invalid",
		"ft",
		"m",
		"",
		"abc123ft",
		"123xyz",
	}

	for _, input := range invalidInputs {
		_, err := parseElevation(input)
		if err == nil {
			t.Errorf("Expected error for invalid input %s, got none", input)
		}
	}
}

func TestClearDatabase(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "homekit_test_db")
	defer os.RemoveAll(tempDir)

	// Create the directory and some test files
	os.MkdirAll(tempDir, 0755)
	testFiles := []string{"keypair", "uuid", "version", "schema"}
	for _, file := range testFiles {
		testFile := filepath.Join(tempDir, file)
		f, err := os.Create(testFile)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Verify files exist before clearing
	for _, file := range testFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Fatalf("Test file %s should exist before clearing", file)
		}
	}

	// Clear the database
	err := ClearDatabase(tempDir)
	if err != nil {
		t.Fatalf("ClearDatabase failed: %v", err)
	}

	// Verify files are gone
	for _, file := range testFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); !os.IsNotExist(err) {
			t.Errorf("Test file %s should not exist after clearing", file)
		}
	}
}

func TestClearDatabaseNonExistentDir(t *testing.T) {
	// Should not error when directory doesn't exist
	err := ClearDatabase("/non/existent/directory")
	if err != nil {
		t.Errorf("ClearDatabase should not error on non-existent directory, got: %v", err)
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
