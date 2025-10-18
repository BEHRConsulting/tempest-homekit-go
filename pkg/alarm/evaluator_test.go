package alarm

import (
	"testing"

	"tempest-homekit-go/pkg/weather"
)

func TestEvaluatorSimpleConditions(t *testing.T) {
	obs := &weather.Observation{
		AirTemperature:       30.0,
		RelativeHumidity:     75.0,
		StationPressure:      1013.25,
		WindAvg:              5.5,
		WindGust:             8.0,
		WindDirection:        180,
		Illuminance:          15000,
		UV:                   7,
		RainAccumulated:      2.5,
		LightningStrikeCount: 3,
		LightningStrikeAvg:   5.2,
		PrecipitationType:    1,
	}

	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		condition string
		want      bool
		wantError bool
	}{
		// Temperature tests
		{"temp greater than", "temperature > 25", true, false},
		{"temp less than", "temperature < 25", false, false},
		{"temp equals", "temperature == 30", true, false},
		{"temp not equals", "temperature != 25", true, false},
		{"temp greater or equal", "temperature >= 30", true, false},
		{"temp less or equal", "temperature <= 30", true, false},
		{"temp alias", "temp > 25", true, false},

		// Humidity tests
		{"humidity greater", "humidity > 70", true, false},
		{"humidity less", "humidity < 70", false, false},

		// Pressure tests
		{"pressure greater", "pressure > 1000", true, false},
		{"pressure less", "pressure < 1000", false, false},

		// Wind tests
		{"wind speed greater", "wind_speed > 5", true, false},
		{"wind alias", "wind > 5", true, false},
		{"wind gust greater", "wind_gust > 7", true, false},
		{"wind direction", "wind_direction == 180", true, false},

		// Light tests
		{"lux greater", "lux > 10000", true, false},
		{"light alias", "light > 10000", true, false},

		// UV tests
		{"uv greater", "uv > 5", true, false},
		{"uv index alias", "uv_index > 5", true, false},

		// Rain tests
		{"rain rate", "rain_rate > 2", true, false},
		{"rain daily", "rain_daily > 2", true, false},

		// Lightning tests
		{"lightning count", "lightning_count > 0", true, false},
		{"lightning distance", "lightning_distance < 10", true, false},

		// Precipitation type
		{"precip type", "precipitation_type == 1", true, false},

		// Invalid field
		{"invalid field", "fake_field > 100", false, true},
		{"invalid operator", "temperature >> 25", false, true},
		{"invalid syntax", "temperature", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.condition, obs)
			if (err != nil) != tt.wantError {
				t.Errorf("Evaluate() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result != tt.want {
				t.Errorf("Evaluate() = %v, want %v for condition: %s", result, tt.want, tt.condition)
			}
		})
	}
}

func TestEvaluatorCompoundConditions(t *testing.T) {
	obs := &weather.Observation{
		AirTemperature:   30.0,
		RelativeHumidity: 85.0,
		Illuminance:      20000,
		UV:               8,
	}

	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		condition string
		want      bool
	}{
		{"and both true", "temperature > 25 && humidity > 80", true},
		{"and first false", "temperature < 25 && humidity > 80", false},
		{"and second false", "temperature > 25 && humidity < 80", false},
		{"and both false", "temperature < 25 && humidity < 80", false},

		{"or both true", "temperature > 25 || humidity > 80", true},
		{"or first true", "temperature > 25 || humidity < 80", true},
		{"or second true", "temperature < 25 || humidity > 80", true},
		{"or both false", "temperature < 25 || humidity < 80", false},

		{"multiple and", "temperature > 25 && humidity > 80 && lux > 15000", true},
		{"multiple or", "temperature < 25 || humidity < 80 || lux > 15000", true},

		{"complex condition", "temperature > 25 && humidity > 80", true},
		{"heat index condition", "temperature > 30 && humidity > 60", false}, // temp is exactly 30
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.condition, obs)
			if err != nil {
				t.Fatalf("Evaluate() unexpected error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Evaluate() = %v, want %v for condition: %s", result, tt.want, tt.condition)
			}
		})
	}
}

func TestEvaluatorGetAvailableFields(t *testing.T) {
	evaluator := NewEvaluator()
	fields := evaluator.GetAvailableFields()

	// Check that we have the expected number of fields
	if len(fields) < 10 {
		t.Errorf("GetAvailableFields() returned %d fields, expected at least 10", len(fields))
	}

	// Check for some key fields
	expectedFields := []string{
		"temperature", "temp",
		"humidity",
		"pressure",
		"wind_speed", "wind",
		"lux", "light",
		"uv", "uv_index",
		"rain_rate",
		"lightning_count",
	}

	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field] = true
	}

	for _, expected := range expectedFields {
		if !fieldMap[expected] {
			t.Errorf("GetAvailableFields() missing expected field: %s", expected)
		}
	}
}

func TestEvaluatorEdgeCases(t *testing.T) {
	obs := &weather.Observation{
		AirTemperature: 0.0,
		UV:             0,
	}

	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		condition string
		want      bool
		wantError bool
	}{
		{"zero temperature equals", "temperature == 0", true, false},
		{"zero temperature not equals", "temperature != 0", false, false},
		{"zero uv", "uv == 0", true, false},
		{"whitespace condition", "  temperature > 0  ", false, false},
		{"empty condition", "", false, true},
		{"only operator", "> 5", false, true},
		{"missing value", "temperature >", false, true},
		{"missing operator", "temperature 5", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.condition, obs)
			if (err != nil) != tt.wantError {
				t.Errorf("Evaluate() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result != tt.want {
				t.Errorf("Evaluate() = %v, want %v for condition: %s", result, tt.want, tt.condition)
			}
		})
	}
}

func TestEvaluatorNegativeValues(t *testing.T) {
	obs := &weather.Observation{
		AirTemperature:  -10.0,
		StationPressure: 980.5,
	}

	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		condition string
		want      bool
	}{
		{"negative temp greater", "temperature > -15", true},
		{"negative temp less", "temperature < -5", true},
		{"negative temp equals", "temperature == -10", true},
		{"pressure comparison", "pressure < 1000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.condition, obs)
			if err != nil {
				t.Fatalf("Evaluate() unexpected error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Evaluate() = %v, want %v for condition: %s", result, tt.want, tt.condition)
			}
		})
	}
}

func TestEvaluatorFloatPrecision(t *testing.T) {
	obs := &weather.Observation{
		AirTemperature: 25.5,
		WindAvg:        3.14159,
	}

	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		condition string
		want      bool
	}{
		{"float comparison greater", "temperature > 25.4", true},
		{"float comparison less", "temperature < 25.6", true},
		{"float comparison equals", "temperature == 25.5", true},
		{"wind decimal", "wind_speed > 3.14", true},
		{"wind decimal precise", "wind_speed < 3.15", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.condition, obs)
			if err != nil {
				t.Fatalf("Evaluate() unexpected error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Evaluate() = %v, want %v for condition: %s", result, tt.want, tt.condition)
			}
		})
	}
}
