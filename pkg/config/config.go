package config

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Token       string
	StationName string
	Pin         string
	LogLevel    string
	WebPort     string
	ClearDB     bool
}

func LoadConfig() *Config {
	cfg := &Config{
		Token:       getEnvOrDefault("TEMPEST_TOKEN", "b88edc78-6261-414e-8042-86a4d4f9ba15"),
		StationName: getEnvOrDefault("TEMPEST_STATION_NAME", "Chino Hills"),
		Pin:         getEnvOrDefault("HOMEKIT_PIN", "00102003"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "error"),
		WebPort:     getEnvOrDefault("WEB_PORT", "8080"),
	}

	flag.StringVar(&cfg.Token, "token", cfg.Token, "WeatherFlow API token")
	flag.StringVar(&cfg.StationName, "station", cfg.StationName, "Tempest station name")
	flag.StringVar(&cfg.Pin, "pin", cfg.Pin, "HomeKit PIN")
	flag.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel, "Log level (debug, info, error)")
	flag.StringVar(&cfg.WebPort, "web-port", cfg.WebPort, "Web dashboard port")
	flag.BoolVar(&cfg.ClearDB, "cleardb", false, "Clear HomeKit database and reset device pairing")
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
