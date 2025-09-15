package config

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Token       string
	StationName string
	Pin         string
	LogLevel    string
	WebPort     string
	ClearDB     bool
	Sensors     string
	ReadHistory bool
	TestAPI     bool
}

func LoadConfig() *Config {
	cfg := &Config{
		Token:       getEnvOrDefault("TEMPEST_TOKEN", "b88edc78-6261-414e-8042-86a4d4f9ba15"),
		StationName: getEnvOrDefault("TEMPEST_STATION_NAME", "Chino Hills"),
		Pin:         getEnvOrDefault("HOMEKIT_PIN", "00102003"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "error"),
		WebPort:     getEnvOrDefault("WEB_PORT", "8080"),
		Sensors:     getEnvOrDefault("SENSORS", "min"),
	}

	flag.StringVar(&cfg.Token, "token", cfg.Token, "WeatherFlow API token")
	flag.StringVar(&cfg.StationName, "station", cfg.StationName, "Tempest station name")
	flag.StringVar(&cfg.Pin, "pin", cfg.Pin, "HomeKit PIN")
	flag.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel, "Log level (debug, info, error)")
	flag.StringVar(&cfg.WebPort, "web-port", cfg.WebPort, "Web dashboard port")
	flag.StringVar(&cfg.Sensors, "sensors", cfg.Sensors, "Sensors to enable: 'all', 'min', 'temp-only', or comma-delimited list (temp,humidity,lux,wind,rain,pressure)")
	flag.BoolVar(&cfg.ClearDB, "cleardb", false, "Clear HomeKit database and reset device pairing")
	flag.BoolVar(&cfg.ReadHistory, "read-history", false, "Preload last 24 hours of weather data from Tempest API")
	flag.BoolVar(&cfg.TestAPI, "test-api", false, "Test WeatherFlow API endpoints and data points")
	flag.Parse()

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

// ParseSensorConfig parses the sensor configuration string
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
			// Only temperature sensor is HomeKit compliant - others cause "out of compliance" errors
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
