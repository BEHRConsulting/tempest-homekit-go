package config

import (
	"flag"
	"os"
)

type Config struct {
	Token       string
	StationName string
	Pin         string
	LogLevel    string
	WebPort     string
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
	flag.Parse()

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
