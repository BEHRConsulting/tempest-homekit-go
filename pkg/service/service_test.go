package service

import (
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	// Test that setLogLevel doesn't panic with various inputs
	testLevels := []string{"debug", "info", "error", "invalid"}

	for _, level := range testLevels {
		// Should not panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("setLogLevel(%s) panicked: %v", level, r)
				}
			}()
			setLogLevel(level)
		}()
	}
}

func TestIsNightTime(t *testing.T) {
	tests := []struct {
		illuminance float64
		expected    bool
		description string
	}{
		{5.0, true, "Low illuminance should be night"},
		{50.0, false, "High illuminance should be day"},
		{9.9, true, "Just below threshold should be night"},
		{10.0, false, "At threshold should be day"},
		{0.0, true, "Zero illuminance should be night"},
	}

	for _, test := range tests {
		result := isNightTime(test.illuminance)
		if result != test.expected {
			t.Errorf("isNightTime(%f) = %t, expected %t (%s)",
				test.illuminance, result, test.expected, test.description)
		}
	}
}
