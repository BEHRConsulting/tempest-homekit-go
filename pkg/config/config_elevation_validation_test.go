package config

import (
	"strings"
	"testing"
)

// TestParseElevationValidation tests elevation range validation with proper error messages
func TestParseElevationValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldError   bool
		errorContains string
	}{
		{
			name:        "Death Valley elevation",
			input:       "-282ft",
			shouldError: false,
		},
		{
			name:        "Dead Sea elevation",
			input:       "-430m",
			shouldError: false,
		},
		{
			name:        "Mount Everest elevation",
			input:       "29029ft",
			shouldError: false,
		},
		{
			name:        "Sea level",
			input:       "0ft",
			shouldError: false,
		},
		{
			name:          "Below Dead Sea",
			input:         "-500m",
			shouldError:   true,
			errorContains: "below Earth's lowest point",
		},
		{
			name:          "Above Mount Everest",
			input:         "9000m",
			shouldError:   true,
			errorContains: "above Earth's highest point",
		},
		{
			name:          "Way too low in feet",
			input:         "-2000ft",
			shouldError:   true,
			errorContains: "below Earth's lowest point",
		},
		{
			name:          "Way too high in feet",
			input:         "35000ft",
			shouldError:   true,
			errorContains: "above Earth's highest point",
		},
		{
			name:          "Airplane cruising altitude (too high)",
			input:         "35000ft",
			shouldError:   true,
			errorContains: "above Earth's highest point",
		},
		{
			name:          "Ocean trench depth (too low)",
			input:         "-11000m",
			shouldError:   true,
			errorContains: "below Earth's lowest point",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseElevation(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for elevation '%s', but got none", tt.input)
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s' for elevation '%s', got: %v",
						tt.errorContains, tt.input, err)
				}
				// Result should be 0 for errors
				if result != 0 {
					t.Errorf("Expected result 0 for error case, got %f", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for elevation '%s', but got: %v", tt.input, err)
				}
				// Result should be reasonable for valid cases
				if result < -430.1 || result > 8848.1 {
					t.Errorf("Result %f for elevation '%s' is outside expected range", result, tt.input)
				}
			}
		})
	}
}

// TestElevationRealWorldExamples tests real-world elevation examples
func TestElevationRealWorldExamples(t *testing.T) {
	realWorldElevations := []struct {
		name     string
		input    string
		location string
		valid    bool
	}{
		{"Death Valley, California", "-282ft", "Lowest point in North America", true},
		{"Dead Sea shore", "-430m", "Lowest land point on Earth", true},
		{"Denver, Colorado", "5280ft", "Mile High City", true},
		{"Mount Everest summit", "29029ft", "Highest point on Earth", true},
		{"Mount McKinley/Denali", "20310ft", "Highest peak in North America", true},
		{"Challenger Deep", "-36200ft", "Deepest ocean trench (invalid for land elevation)", false},
		{"Commercial jet altitude", "35000ft", "Too high for land elevation", false},
		{"Space station", "400000m", "Way too high", false},
		{"Mariana Trench", "-11000m", "Ocean depth (not land elevation)", false},
	}

	for _, example := range realWorldElevations {
		t.Run(example.name, func(t *testing.T) {
			_, err := parseElevation(example.input)

			if example.valid && err != nil {
				t.Errorf("Expected %s elevation '%s' to be valid, but got error: %v",
					example.name, example.input, err)
			}
			if !example.valid && err == nil {
				t.Errorf("Expected %s elevation '%s' to be invalid, but got no error",
					example.name, example.input)
			}
		})
	}
}
