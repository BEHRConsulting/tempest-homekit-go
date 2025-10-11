package config

import (
	"strings"
	"testing"
)

// TestValidateConfigValid tests that valid configurations pass validation
func TestValidateConfigValid(t *testing.T) {
	validCfg := &Config{
		Token:       "valid-token",
		StationName: "Test Station",
		Pin:         "12345678",
		LogLevel:    "debug",
		WebPort:     "8080",
		Sensors:     "temp,humidity,lux",
	}

	if err := validateConfig(validCfg); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
}

// TestValidateConfigInvalidLogLevel tests log level validation
func TestValidateConfigInvalidLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{"invalid level", "invalid"},
		{"empty level", ""},
		{"uppercase", "DEBUG"},
		{"mixed case", "Info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    tt.logLevel,
				WebPort:     "8080",
				Sensors:     "temp",
			}

			err := validateConfig(cfg)
			if err == nil {
				t.Errorf("Expected invalid log level '%s' to fail validation", tt.logLevel)
			}
			if !strings.Contains(err.Error(), "invalid log level") {
				t.Errorf("Expected log level error, got: %v", err)
			}
		})
	}
}

// TestValidateConfigValidLogLevels tests all valid log levels
func TestValidateConfigValidLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "warning", "error"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    level,
				WebPort:     "8080",
				Sensors:     "temp",
			}

			if err := validateConfig(cfg); err != nil {
				t.Errorf("Expected valid log level '%s' to pass, got error: %v", level, err)
			}
		})
	}
}

// TestValidateConfigInvalidSensors tests sensor validation
func TestValidateConfigInvalidSensors(t *testing.T) {
	tests := []struct {
		name    string
		sensors string
	}{
		{"invalid sensor", "invalid-sensor"},
		{"mixed valid/invalid", "temp,invalid,humidity"},
		{"unknown preset", "unknown-preset"},
		{"typo in sensor", "temprature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    "debug",
				WebPort:     "8080",
				Sensors:     tt.sensors,
			}

			err := validateConfig(cfg)
			if err == nil {
				t.Errorf("Expected invalid sensors '%s' to fail validation", tt.sensors)
			}
			if !strings.Contains(err.Error(), "invalid sensor") {
				t.Errorf("Expected sensor error, got: %v", err)
			}
		})
	}
}

// TestValidateConfigValidSensors tests valid sensor configurations
func TestValidateConfigValidSensors(t *testing.T) {
	validSensors := []string{
		"all",
		"min",
		"temp",
		"temperature",
		"temp,humidity",
		"temp,humidity,lux",
		"temperature,humidity,light",
		"temp,humidity,lux,wind,rain,pressure,uv,lightning",
		"temperature,humidity,light,wind,rain,pressure,uvi,lightning",
		"light", // alias for lux
		"uvi",   // alias for uv
	}

	for _, sensors := range validSensors {
		t.Run(sensors, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    "debug",
				WebPort:     "8080",
				Sensors:     sensors,
			}

			if err := validateConfig(cfg); err != nil {
				t.Errorf("Expected valid sensors '%s' to pass, got error: %v", sensors, err)
			}
		})
	}
}

// TestValidateConfigInvalidWebPort tests web port validation
func TestValidateConfigInvalidWebPort(t *testing.T) {
	tests := []struct {
		name    string
		webPort string
	}{
		{"non-numeric", "not-a-number"},
		{"empty", ""},
		{"decimal", "8080.5"},
		{"letters mixed", "80a80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    "debug",
				WebPort:     tt.webPort,
				Sensors:     "temp",
			}

			err := validateConfig(cfg)
			if err == nil {
				t.Errorf("Expected invalid web port '%s' to fail validation", tt.webPort)
			}
			if !strings.Contains(err.Error(), "invalid web port") {
				t.Errorf("Expected web port error, got: %v", err)
			}
		})
	}
}

// TestValidateConfigValidWebPorts tests valid web port values
func TestValidateConfigValidWebPorts(t *testing.T) {
	validPorts := []string{"8080", "3000", "80", "443", "8000", "9000"}

	for _, port := range validPorts {
		t.Run(port, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    "debug",
				WebPort:     port,
				Sensors:     "temp",
			}

			if err := validateConfig(cfg); err != nil {
				t.Errorf("Expected valid web port '%s' to pass, got error: %v", port, err)
			}
		})
	}
}

// TestValidateConfigInvalidPin tests PIN validation
func TestValidateConfigInvalidPin(t *testing.T) {
	tests := []struct {
		name string
		pin  string
	}{
		{"too short", "123"},
		{"too long", "123456789"},
		{"empty", ""},
		{"contains letters", "1234567a"},
		{"contains symbols", "1234567!"},
		{"seven digits", "1234567"},
		{"nine digits", "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         tt.pin,
				LogLevel:    "debug",
				WebPort:     "8080",
				Sensors:     "temp",
			}

			err := validateConfig(cfg)
			if err == nil {
				t.Errorf("Expected invalid PIN '%s' to fail validation", tt.pin)
			}
			if !strings.Contains(err.Error(), "invalid HomeKit PIN") {
				t.Errorf("Expected PIN error, got: %v", err)
			}
		})
	}
}

// TestValidateConfigValidPins tests valid PIN formats
func TestValidateConfigValidPins(t *testing.T) {
	validPins := []string{"12345678", "00000000", "99999999", "12340000"}

	for _, pin := range validPins {
		t.Run(pin, func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         pin,
				LogLevel:    "debug",
				WebPort:     "8080",
				Sensors:     "temp",
			}

			if err := validateConfig(cfg); err != nil {
				t.Errorf("Expected valid PIN '%s' to pass, got error: %v", pin, err)
			}
		})
	}
}

// TestValidateConfigEmptyRequiredFields tests required field validation
func TestValidateConfigEmptyRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		stationName string
		expectError string
	}{
		{"empty token", "", "Test Station", "WeatherFlow API token is required"},
		{"empty station name", "valid-token", "", "station name is required"},
		{"both empty", "", "", "WeatherFlow API token is required"}, // Should catch token first
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Token:       tt.token,
				StationName: tt.stationName,
				Pin:         "12345678",
				LogLevel:    "debug",
				WebPort:     "8080",
				Sensors:     "temp",
			}

			err := validateConfig(cfg)
			if err == nil {
				t.Errorf("Expected validation to fail for %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectError, err)
			}
		})
	}
}

// TestValidateConfigEmptySensors tests that empty sensors is allowed
func TestValidateConfigEmptySensors(t *testing.T) {
	cfg := &Config{
		Token:       "valid-token",
		StationName: "Test Station",
		Pin:         "12345678",
		LogLevel:    "debug",
		WebPort:     "8080",
		Sensors:     "",
	}

	// Empty sensors should be valid (will use defaults)
	if err := validateConfig(cfg); err != nil {
		t.Errorf("Expected empty sensors to be valid, got error: %v", err)
	}
}

// TestValidateConfigSensorsWithSpaces tests sensor parsing with various spacing
func TestValidateConfigSensorsWithSpaces(t *testing.T) {
	validCfgs := []string{
		"temp, humidity, lux",
		" temp,humidity,lux ",
		"temp ,humidity , lux",
		"temp  ,  humidity  ,  lux",
	}

	for _, sensors := range validCfgs {
		t.Run("'"+sensors+"'", func(t *testing.T) {
			cfg := &Config{
				Token:       "valid-token",
				StationName: "Test Station",
				Pin:         "12345678",
				LogLevel:    "debug",
				WebPort:     "8080",
				Sensors:     sensors,
			}

			if err := validateConfig(cfg); err != nil {
				t.Errorf("Expected sensors with spaces '%s' to be valid, got error: %v", sensors, err)
			}
		})
	}
}

// TestValidateConfigDisableInternetRequiresUDPStream tests that --disable-internet requires a local data source
func TestValidateConfigDisableInternetRequiresUDPStream(t *testing.T) {
	cfg := &Config{
		Token:               "valid-token",
		StationName:         "Test Station",
		Pin:                 "12345678",
		LogLevel:            "debug",
		WebPort:             "8080",
		Sensors:             "temp",
		DisableInternet:     true,
		UDPStream:           false,
		UseGeneratedWeather: false,
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("Expected --disable-internet without data source to fail validation")
	}
	if !strings.Contains(err.Error(), "--disable-internet mode requires") {
		t.Errorf("Expected data source requirement error, got: %v", err)
	}
}

// TestValidateConfigDisableInternetWithUseWebStatus tests that --disable-internet rejects --use-web-status
func TestValidateConfigDisableInternetWithUseWebStatus(t *testing.T) {
	cfg := &Config{
		Token:           "valid-token",
		StationName:     "Test Station",
		Pin:             "12345678",
		LogLevel:        "debug",
		WebPort:         "8080",
		Sensors:         "temp",
		DisableInternet: true,
		UDPStream:       true,
		UseWebStatus:    true,
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("Expected --disable-internet with --use-web-status to fail validation")
	}
	if !strings.Contains(err.Error(), "--use-web-status cannot be used with --disable-internet") {
		t.Errorf("Expected web status conflict error, got: %v", err)
	}
}

// TestValidateConfigDisableInternetWithReadHistory tests that --disable-internet rejects --read-history
func TestValidateConfigDisableInternetWithReadHistory(t *testing.T) {
	cfg := &Config{
		Token:           "valid-token",
		StationName:     "Test Station",
		Pin:             "12345678",
		LogLevel:        "debug",
		WebPort:         "8080",
		Sensors:         "temp",
		DisableInternet: true,
		UDPStream:       true,
		ReadHistory:     true,
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("Expected --disable-internet with --read-history to fail validation")
	}
	if !strings.Contains(err.Error(), "--read-history cannot be used with --disable-internet") {
		t.Errorf("Expected read history conflict error, got: %v", err)
	}
}

// TestValidateConfigDisableInternetValid tests that --disable-internet with --udp-stream is valid
func TestValidateConfigDisableInternetValid(t *testing.T) {
	cfg := &Config{
		Token:           "valid-token",
		StationName:     "Test Station",
		Pin:             "12345678",
		LogLevel:        "debug",
		WebPort:         "8080",
		Sensors:         "temp",
		DisableInternet: true,
		UDPStream:       true,
	}

	if err := validateConfig(cfg); err != nil {
		t.Errorf("Expected valid --disable-internet config to pass, got error: %v", err)
	}
}

// TestValidateConfigDisableInternetWithGeneratedWeather tests that --disable-internet with --use-generated-weather is valid
func TestValidateConfigDisableInternetWithGeneratedWeather(t *testing.T) {
	cfg := &Config{
		Token:               "valid-token",
		StationName:         "Test Station",
		Pin:                 "12345678",
		LogLevel:            "debug",
		WebPort:             "8080",
		Sensors:             "temp",
		DisableInternet:     true,
		UseGeneratedWeather: true,
	}

	if err := validateConfig(cfg); err != nil {
		t.Errorf("Expected valid --disable-internet with --use-generated-weather config to pass, got error: %v", err)
	}
}

// TestValidateConfigDisableInternetRequiresDataSource tests that --disable-internet requires either UDP or generated weather
func TestValidateConfigDisableInternetRequiresDataSource(t *testing.T) {
	cfg := &Config{
		Token:               "valid-token",
		StationName:         "Test Station",
		Pin:                 "12345678",
		LogLevel:            "debug",
		WebPort:             "8080",
		Sensors:             "temp",
		DisableInternet:     true,
		UDPStream:           false,
		UseGeneratedWeather: false,
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("Expected --disable-internet without data source to fail validation")
	}
	if !strings.Contains(err.Error(), "requires --udp-stream or --use-generated-weather") {
		t.Errorf("Expected data source requirement error, got: %v", err)
	}
}
