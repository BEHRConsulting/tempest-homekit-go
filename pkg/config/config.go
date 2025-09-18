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
	Token        string
	StationName  string
	Pin          string
	LogLevel     string
	WebPort      string
	ClearDB      bool
	Sensors      string
	ReadHistory  bool
	TestAPI      bool
	UseWebStatus bool // Enable headless browser scraping of TempestWX status
	Elevation    float64 // elevation in meters
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
		Sensors:     getEnvOrDefault("SENSORS", "temp,lux,humidity"),
		Elevation:   275.2, // 903ft default elevation in meters
	}

	var elevationStr string
	var elevationProvided bool
	flag.StringVar(&cfg.Token, "token", cfg.Token, "WeatherFlow API token")
	flag.StringVar(&cfg.StationName, "station", cfg.StationName, "Tempest station name")
	flag.StringVar(&cfg.Pin, "pin", cfg.Pin, "HomeKit PIN")
	flag.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel, "Log level (debug, info, error)")
	flag.StringVar(&cfg.WebPort, "web-port", cfg.WebPort, "Web dashboard port")
	flag.StringVar(&cfg.Sensors, "sensors", cfg.Sensors, "Sensors to enable: 'all', 'min' (temp,lux,humidity), 'temp-only', or comma-delimited list (temp,humidity,lux,wind,rain,pressure)")
	flag.StringVar(&elevationStr, "elevation", "", "Station elevation (e.g., 903ft, 275m). If not provided, elevation will be auto-detected from coordinates")
	flag.BoolVar(&cfg.ClearDB, "cleardb", false, "Clear HomeKit database and reset device pairing")
	flag.BoolVar(&cfg.ReadHistory, "read-history", false, "Preload last 24 hours of weather data from Tempest API")
	flag.BoolVar(&cfg.TestAPI, "test-api", false, "Test WeatherFlow API endpoints and data points")
	flag.BoolVar(&cfg.UseWebStatus, "use-web-status", false, "Enable headless browser scraping of TempestWX status page every 15 minutes")

	// Parse flags but check if elevation was actually provided
	flag.Parse()

	// Check if elevation was provided by user
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "elevation" {
			elevationProvided = true
		}
	})

	// Handle elevation configuration - auto lookup by default
	if !elevationProvided || strings.ToLower(elevationStr) == "auto" {
		if elevation, err := lookupStationElevation(cfg.Token, cfg.StationName); err != nil {
			log.Printf("Warning: Failed to lookup elevation automatically: %v", err)
			log.Printf("Using fallback elevation 903ft (275.2m)")
		} else {
			cfg.Elevation = elevation
			log.Printf("Auto-detected elevation: %.1f meters (%.0f feet)", elevation, elevation*3.28084)
		}
	} else {
		// Parse manually provided elevation with units
		if elevation, err := parseElevation(elevationStr); err != nil {
			log.Printf("Warning: Invalid elevation format '%s', using fallback 903ft (275.2m): %v", elevationStr, err)
		} else {
			cfg.Elevation = elevation
			log.Printf("Using specified elevation: %.1f meters (%.0f feet)", elevation, elevation*3.28084)
		}
	}

	return cfg
}

// ClearDatabase removes all files in the HomeKit database directory
func ClearDatabase(dbPath string) error {
	log.Printf("Clearing HomeKit database at: %s", dbPath)

	// Check if directory exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("Database directory does not exist: %s", dbPath)
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
			log.Printf("Removed: %s", filepath.Base(file))
		}
	}

	log.Printf("HomeKit database cleared successfully")
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
// Supported values: "all", "min", "temp-only", or comma-separated sensor names.
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
			// Core sensors: temperature, humidity, and lux for comprehensive weather monitoring
		}
	case "temp-only":
		return SensorConfig{
			Temperature: true,
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
			case "uv":
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

	if strings.HasSuffix(elevationStr, "ft") {
		// Parse feet and convert to meters
		valueStr := strings.TrimSuffix(elevationStr, "ft")
		feet, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, err
		}
		return feet * 0.3048, nil // 1 foot = 0.3048 meters
	} else if strings.HasSuffix(elevationStr, "m") {
		// Parse meters directly
		valueStr := strings.TrimSuffix(elevationStr, "m")
		meters, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, err
		}
		return meters, nil
	} else {
		// Try to parse as number without unit, assume meters
		meters, err := strconv.ParseFloat(elevationStr, 64)
		if err != nil {
			return 0, err
		}
		return meters, nil
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
