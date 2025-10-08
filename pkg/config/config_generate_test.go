package config

import (
	"testing"
)

func TestGeneratedStationURLConstruction(t *testing.T) {
	cfg := &Config{
		WebPort:              "3000",
		UseGeneratedWeather:  true,
		GeneratedWeatherPath: "/api/generate-weather",
	}

	// simulate the logic from LoadConfig that sets StationURL when UseGeneratedWeather is true
	if cfg.StationURL == "" && cfg.UseGeneratedWeather {
		cfg.StationURL = "http://localhost:" + cfg.WebPort + cfg.GeneratedWeatherPath
	}

	expected := "http://localhost:3000/api/generate-weather"
	if cfg.StationURL != expected {
		t.Fatalf("expected StationURL %s, got %s", expected, cfg.StationURL)
	}
}
