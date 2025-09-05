package config

import (
	"os"
	"testing"
)

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
