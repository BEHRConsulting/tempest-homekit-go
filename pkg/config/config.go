// Package config provides configuration management for the Tempest HomeKit service.
// It handles command-line flags, environment variables, and HomeKit database operations.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration parameters for the Tempest HomeKit service.
type Config struct {
	Token               string
	StationName         string
	Pin                 string
	LogLevel            string
	WebPort             string
	ClearDB             bool
	Sensors             string
	ReadHistory         bool
	TestAPI             bool
	UseWebStatus        bool    // Enable headless browser scraping of TempestWX status
	UseGeneratedWeather bool    // Use generated weather data for testing instead of Tempest API
	Elevation           float64 // elevation in meters
	Version             bool    // Show version and exit
}

// LoadConfig initializes and returns a new Config struct with values from
// environment variables, command-line flags, and sensible defaults.
func LoadConfig() *Config {
	cfg := &Config{
		Token:       getEnvOrDefault("TEMPEST_TOKEN", "b88edc78-6261-414e-8042-86a4d4f9ba15"),
		StationName: getEnvOrDefault("TEMPEST_STATION_NAME", "Chino Hills"),
		Pin:         getEnvOrDefault("HOMEKIT_PIN", "00102003"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "error"),
		WebPort:     getEnvOrDefault("WEB_PORT", "8080"),
		Sensors:     getEnvOrDefault("SENSORS", "temp,lux,humidity,uv"),
		Elevation:   275.2, // 903ft default elevation in meters
	}

	var elevationStr string
	var elevationProvided bool
	flag.StringVar(&cfg.Token, "token", cfg.Token, "WeatherFlow API token")
	flag.StringVar(&cfg.StationName, "station", cfg.StationName, "Tempest station name")
	flag.StringVar(&cfg.Pin, "pin", cfg.Pin, "HomeKit PIN")
	flag.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel, "Log level (debug, info, error)")
	flag.StringVar(&cfg.WebPort, "web-port", cfg.WebPort, "Web dashboard port")
	flag.StringVar(&cfg.Sensors, "sensors", cfg.Sensors, "Sensors to enable: 'all', 'min' (temp,humidity,lux), or comma-delimited list (temp/temperature,humidity,lux/light,wind,rain,pressure,uv/uvi,lightning)")
	flag.StringVar(&elevationStr, "elevation", "", "Station elevation (e.g., 903ft, 275m). If not provided, elevation will be auto-detected from coordinates")
	flag.BoolVar(&cfg.ClearDB, "cleardb", false, "Clear HomeKit database and reset device pairing")
	flag.BoolVar(&cfg.ReadHistory, "read-history", false, "Preload last 24 hours of weather data from Tempest API")
	flag.BoolVar(&cfg.TestAPI, "test-api", false, "Test WeatherFlow API endpoints and data points")
	flag.BoolVar(&cfg.UseWebStatus, "use-web-status", false, "Enable headless browser scraping of TempestWX status page every 15 minutes")
	flag.BoolVar(&cfg.UseGeneratedWeather, "use-generated-weather", false, "Use generated weather data for UI testing instead of Tempest API")
	flag.BoolVar(&cfg.Version, "version", false, "Show version information and exit")

	// Parse flags but check if elevation was actually provided
	flag.Parse()

	// Validate command line arguments
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n\n", err)
		flag.Usage()
		os.Exit(2)
	}

	// Check if elevation was provided by user
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "elevation" {
			elevationProvided = true
		}
	})

	// Handle elevation configuration - auto lookup by default
	if !elevationProvided || strings.ToLower(elevationStr) == "auto" {
		// Skip station elevation lookup if using generated weather - elevation will be set later from generated location
		if !cfg.UseGeneratedWeather {
			if elevation, err := lookupStationElevation(cfg.Token, cfg.StationName); err != nil {
				log.Printf("Warning: Failed to lookup elevation automatically: %v", err)
				log.Printf("INFO: Using fallback elevation 903ft (275.2m)")
			} else {
				cfg.Elevation = elevation
				// Don't log here - will be logged later in main.go after logger is set up
			}
		}
		// For generated weather, elevation will be set by the service from the generated location
	} else {
		// Parse manually provided elevation with units
		if elevation, err := parseElevation(elevationStr); err != nil {
			log.Printf("Warning: Invalid elevation format '%s', using fallback 903ft (275.2m): %v", elevationStr, err)
		} else {
			cfg.Elevation = elevation
			log.Printf("INFO: Using specified elevation: %.1f meters (%.0f feet)", elevation, elevation*3.28084)
		}
	}

	return cfg
}

// validateConfig validates command line arguments and returns an error if invalid
func validateConfig(cfg *Config) error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if cfg.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level '%s'. Valid options: debug, info, error", cfg.LogLevel)
	}

	// Validate sensor configuration by testing parsing
	if cfg.Sensors != "" {
		// Test if sensor config is valid by attempting to parse it
		// This will help catch invalid sensor names early
		validSensorNames := []string{"temp", "temperature", "humidity", "lux", "light", "wind", "rain", "pressure", "uv", "uvi", "lightning"}
		validPresets := []string{"all", "min"}

		// Check if it's a preset
		isPreset := false
		for _, preset := range validPresets {
			if cfg.Sensors == preset {
				isPreset = true
				break
			}
		}

		if !isPreset {
			// Parse comma-separated sensor list
			sensors := strings.Split(strings.ToLower(cfg.Sensors), ",")
			for _, sensor := range sensors {
				sensor = strings.TrimSpace(sensor)
				if sensor == "" {
					continue
				}
				valid := false
				for _, validName := range validSensorNames {
					if sensor == validName {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("invalid sensor '%s'. Valid sensors: %s. Valid presets: %s",
						sensor, strings.Join(validSensorNames, ", "), strings.Join(validPresets, ", "))
				}
			}
		}
	}

	// Validate web port is numeric
	if _, err := strconv.Atoi(cfg.WebPort); err != nil {
		return fmt.Errorf("invalid web port '%s'. Port must be a number", cfg.WebPort)
	}

	// Validate HomeKit PIN format (8 digits)
	if len(cfg.Pin) != 8 {
		return fmt.Errorf("invalid HomeKit PIN '%s'. PIN must be exactly 8 digits", cfg.Pin)
	}
	if _, err := strconv.Atoi(cfg.Pin); err != nil {
		return fmt.Errorf("invalid HomeKit PIN '%s'. PIN must contain only digits", cfg.Pin)
	}

	// Validate required fields
	if cfg.Token == "" {
		return fmt.Errorf("WeatherFlow API token is required. Set via --token flag or TEMPEST_TOKEN environment variable")
	}
	if cfg.StationName == "" {
		return fmt.Errorf("station name is required. Set via --station flag or TEMPEST_STATION_NAME environment variable")
	}

	return nil
}

// ClearDatabase removes all files in the HomeKit database directory
func ClearDatabase(dbPath string) error {
	log.Printf("INFO: Clearing HomeKit database at: %s", dbPath)

	// Check if directory exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("INFO: Database directory does not exist: %s", dbPath)
		return nil
	}

	// Remove all files in the directory
	files, err := filepath.Glob(filepath.Join(dbPath, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			log.Printf("Warning: Failed to remove %s: %v", file, err)
		} else {
			log.Printf("INFO: Removed: %s", filepath.Base(file))
		}
	}

	log.Printf("INFO: HomeKit database cleared successfully")
	return nil
}

// SensorConfig represents which sensors should be enabled
type SensorConfig struct {
	Temperature bool
	Humidity    bool
	Light       bool
	Wind        bool
	Rain        bool
	Pressure    bool
	UV          bool
	Lightning   bool
}

// ParseSensorConfig parses the sensor configuration string and returns a SensorConfig
// with appropriate sensor types enabled based on the input string.
// Supported values: "all", "min", or comma-separated sensor names.
func ParseSensorConfig(sensorsFlag string) SensorConfig {
	switch strings.ToLower(sensorsFlag) {
	case "all":
		return SensorConfig{
			Temperature: true,
			Humidity:    true,
			Light:       true,
			Wind:        true,
			Rain:        true,
			Pressure:    true,
			UV:          true,
			Lightning:   true,
		}
	case "min":
		return SensorConfig{
			Temperature: true,
			Humidity:    true,
			Light:       true,
			// Minimal sensors: temperature, humidity, and light for basic weather monitoring
		}
	default:
		// Parse comma-delimited sensor list
		sensors := strings.Split(strings.ToLower(sensorsFlag), ",")
		config := SensorConfig{}
		for _, sensor := range sensors {
			sensor = strings.TrimSpace(sensor)
			switch sensor {
			case "temp", "temperature":
				config.Temperature = true
			case "humidity":
				config.Humidity = true
			case "light", "lux":
				config.Light = true
			case "wind":
				config.Wind = true
			case "rain":
				config.Rain = true
			case "pressure":
				config.Pressure = true
			case "uv", "uvi":
				config.UV = true
			case "lightning":
				config.Lightning = true
			}
		}
		return config
	}
}

// StationLocation represents station coordinates from WeatherFlow API
type StationLocation struct {
	StationID int     `json:"station_id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Elevation float64 `json:"elevation,omitempty"` // May be provided directly
}

// ElevationResponse represents response from elevation API
type ElevationResponse struct {
	Results []struct {
		Elevation float64 `json:"elevation"`
	} `json:"results"`
}

// lookupStationElevation attempts to get elevation from station coordinates
func lookupStationElevation(token, stationName string) (float64, error) {
	// First try to get station coordinates from WeatherFlow API
	lat, lon, err := getStationCoordinates(token, stationName)
	if err != nil {
		return 0, fmt.Errorf("failed to get station coordinates: %v", err)
	}

	// Then lookup elevation from coordinates
	elevation, err := getElevationFromCoordinates(lat, lon)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup elevation for coordinates (%.4f, %.4f): %v", lat, lon, err)
	}

	return elevation, nil
}

// getStationCoordinates fetches station coordinates from WeatherFlow API
func getStationCoordinates(token, stationName string) (lat, lon float64, err error) {
	// First try to get actual station coordinates from WeatherFlow API
	if coords, err := fetchWeatherFlowStationCoords(token, stationName); err == nil {
		return coords[0], coords[1], nil
	}

	// Fallback to known coordinates for common locations
	knownLocations := map[string][2]float64{
		"Chino Hills": {33.9898, -117.7326},
		"Los Angeles": {34.0522, -118.2437},
		"San Diego":   {32.7157, -117.1611},
		"Phoenix":     {33.4484, -112.0740},
		"Denver":      {39.7392, -104.9903},
		"Seattle":     {47.6062, -122.3321},
		"Portland":    {45.5152, -122.6784},
		"Austin":      {30.2672, -97.7431},
		"Dallas":      {32.7767, -96.7970},
		"Miami":       {25.7617, -80.1918},
	}

	if coords, found := knownLocations[stationName]; found {
		return coords[0], coords[1], nil
	}

	return 0, 0, fmt.Errorf("coordinates not available for station '%s' (consider adding coordinates to known locations)", stationName)
}

// fetchWeatherFlowStationCoords attempts to get coordinates from WeatherFlow API
func fetchWeatherFlowStationCoords(_token, _stationName string) (coords [2]float64, err error) {
	// Explicitly ignore unused parameters to satisfy linter
	_ = _token
	_ = _stationName

	// This would query the WeatherFlow API stations endpoint for detailed station info
	// The API might have an endpoint like: /stations/:station_id/details that includes lat/lon
	// For now, we return an error to fall back to known locations

	// TODO: Implement actual WeatherFlow API call when station details endpoint is available
	// Example implementation would be:
	/*
		url := fmt.Sprintf("https://swd.weatherflow.com/swd/rest/stations/%s/details?token=%s", stationID, token)
		resp, err := http.Get(url)
		if err != nil {
			return coords, err
		}
		defer resp.Body.Close()

		var stationDetails StationDetailsResponse
		if err := json.NewDecoder(resp.Body).Decode(&stationDetails); err != nil {
			return coords, err
		}

		if len(stationDetails.Stations) > 0 {
			station := stationDetails.Stations[0]
			coords[0] = station.Latitude
			coords[1] = station.Longitude
			return coords, nil
		}
	*/

	return coords, fmt.Errorf("WeatherFlow station coordinates API not implemented")
}

// getElevationFromCoordinates uses Open Elevation API to get elevation
func getElevationFromCoordinates(lat, lon float64) (float64, error) {
	// Use Open Elevation API (free, no API key required)
	url := fmt.Sprintf("https://api.open-elevation.com/api/v1/lookup?locations=%.4f,%.4f", lat, lon)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("elevation API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("elevation API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read elevation API response: %v", err)
	}

	var elevResp ElevationResponse
	if err := json.Unmarshal(body, &elevResp); err != nil {
		return 0, fmt.Errorf("failed to parse elevation API response: %v", err)
	}

	if len(elevResp.Results) == 0 {
		return 0, fmt.Errorf("no elevation data returned")
	}

	return elevResp.Results[0].Elevation, nil
}

// parseElevation parses elevation string with units (e.g., "903ft", "275m") and returns meters
func parseElevation(elevationStr string) (float64, error) {
	elevationStr = strings.TrimSpace(strings.ToLower(elevationStr))

	var meters float64
	var err error

	if strings.HasSuffix(elevationStr, "ft") {
		// Parse feet and convert to meters
		valueStr := strings.TrimSuffix(elevationStr, "ft")
		feet, parseErr := strconv.ParseFloat(valueStr, 64)
		if parseErr != nil {
			return 0, parseErr
		}
		meters = feet * 0.3048 // 1 foot = 0.3048 meters
	} else if strings.HasSuffix(elevationStr, "m") {
		// Parse meters directly
		valueStr := strings.TrimSuffix(elevationStr, "m")
		meters, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, err
		}
	} else {
		// Try to parse as number without unit, assume meters
		meters, err = strconv.ParseFloat(elevationStr, 64)
		if err != nil {
			return 0, err
		}
	}

	// Validate elevation range: -1411ft to 29029ft (-430m to 8848m)
	// Dead Sea area is the lowest at -430m, Mount Everest is highest at 8848m
	// Add small tolerance for floating point precision
	const minElevationMeters = -430.1 // -1411 feet with tolerance
	const maxElevationMeters = 8848.1 // 29029 feet with tolerance

	if meters < minElevationMeters {
		return 0, fmt.Errorf("elevation %.1fm is below Earth's lowest point (%.1fm, Dead Sea area)", meters, minElevationMeters)
	}
	if meters > maxElevationMeters {
		return 0, fmt.Errorf("elevation %.1fm is above Earth's highest point (%.1fm, Mount Everest)", meters, maxElevationMeters)
	}

	return meters, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
