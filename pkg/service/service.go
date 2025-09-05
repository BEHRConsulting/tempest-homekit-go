package service

import (
	"fmt"
	"log"
	"os"
	"time"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/homekit"
	"tempest-homekit-go/pkg/weather"
	"tempest-homekit-go/pkg/web"
)

func StartService(cfg *config.Config) error {
	// Set log level
	setLogLevel(cfg.LogLevel)

	log.Println("Starting Tempest HomeKit service...")

	// Get stations
	stations, err := weather.GetStations(cfg.Token)
	if err != nil {
		return fmt.Errorf("failed to get stations: %v", err)
	}

	station := weather.FindStationByName(stations, cfg.StationName)
	if station == nil {
		log.Printf("Available stations:")
		for _, s := range stations {
			log.Printf("  - ID: %d, Name: '%s', StationName: '%s'", s.StationID, s.Name, s.StationName)
		}
		return fmt.Errorf("station '%s' not found", cfg.StationName)
	}

	log.Printf("Found station: %s (ID: %d)", station.Name, station.StationID)

	// Setup HomeKit
	wa := homekit.NewWeatherAccessories()
	t, err := homekit.SetupHomeKit(wa, cfg.Pin)
	if err != nil {
		return fmt.Errorf("failed to setup HomeKit: %v", err)
	}

	// Start HomeKit transport
	go func() {
		log.Println("Starting HomeKit transport...")
		t.Start()
	}()

	// Setup web dashboard
	webServer := web.NewWebServer(cfg.WebPort)
	go func() {
		log.Printf("Starting web dashboard on port %s", cfg.WebPort)
		if err := webServer.Start(); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	// Update HomeKit status in web server
	webServer.UpdateHomeKitStatus(map[string]interface{}{
		"bridge":      true,
		"accessories": 4,
		"pin":         cfg.Pin,
	})

	// Poll weather data
	log.Println("Setting up weather polling...")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Starting polling loop...")

	for range ticker.C {
		obs, err := weather.GetObservation(station.StationID, cfg.Token)
		if err != nil {
			log.Printf("Error getting observation: %v", err)
			continue
		}

		log.Printf("Updating HomeKit: Temp=%.1f°C, Humidity=%.1f%%", obs.AirTemperature, obs.RelativeHumidity)

		// Debug logging for all weather metrics
		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: Weather data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph, Rain: %.3f in",
				obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.RainAccumulated)
		}

		wa.UpdateTemperature(obs.AirTemperature)
		wa.UpdateHumidity(obs.RelativeHumidity)
		wa.UpdateWindSpeed(obs.WindAvg)
		wa.UpdateRainAccumulation(obs.RainAccumulated)

		// Update web dashboard
		webServer.UpdateWeather(obs)
	}
	return nil
}

func setLogLevel(level string) {
	switch level {
	case "debug":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "info":
		log.SetFlags(log.LstdFlags)
	case "error":
		log.SetOutput(os.Stderr)
	default:
		log.SetOutput(os.Stderr)
	}
}
