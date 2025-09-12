package service

import (
	"encoding/json"
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
	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Fetching stations from WeatherFlow API")
	}
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
	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Initializing HomeKit accessories")
	}
	wa := homekit.NewWeatherAccessories()
	t, err := homekit.SetupHomeKit(wa, cfg.Pin)
	if err != nil {
		return fmt.Errorf("failed to setup HomeKit: %v", err)
	}

	// Start HomeKit transport
	go func() {
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			log.Printf("INFO: Starting HomeKit transport with PIN: %s", cfg.Pin)
		}
		t.Start()
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			log.Printf("INFO: HomeKit transport started successfully")
		}
	}()

	// Setup web dashboard
	webServer := web.NewWebServer(cfg.WebPort)
	webServer.SetStationName(station.Name)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Web server panic recovered: %v", r)
			}
		}()
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			log.Printf("INFO: Starting web dashboard on port %s", cfg.WebPort)
		}
		if err := webServer.Start(); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	// Update HomeKit status in web server
	homekitStatus := map[string]interface{}{
		"bridge": true,
		"name":   "Tempest HomeKit Bridge",
		"accessories": len([]string{
			wa.TemperatureAccessory.Info.Name.GetValue(),
			wa.HumidityAccessory.Info.Name.GetValue(),
			wa.WindAccessory.Info.Name.GetValue(),
			wa.WindDirectionAccessory.Info.Name.GetValue(),
			wa.RainAccessory.Info.Name.GetValue(),
			wa.LightAccessory.Info.Name.GetValue(),
		}),
		"accessoryNames": []string{
			wa.TemperatureAccessory.Info.Name.GetValue(),
			wa.HumidityAccessory.Info.Name.GetValue(),
			wa.WindAccessory.Info.Name.GetValue(),
			wa.WindDirectionAccessory.Info.Name.GetValue(),
			wa.RainAccessory.Info.Name.GetValue(),
			wa.LightAccessory.Info.Name.GetValue(),
		},
		"pin": cfg.Pin,
	}
	webServer.UpdateHomeKitStatus(homekitStatus)

	// Poll weather data
	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		log.Printf("INFO: Setting up weather polling every 60 seconds")
	}
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		log.Printf("INFO: Starting weather data polling loop")
	}

	for range ticker.C {
		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: Polling iteration started - fetching observation from station %d", station.StationID)
		}

		obs, err := weather.GetObservation(station.StationID, cfg.Token)
		if err != nil {
			log.Printf("Error getting observation: %v", err)
			continue
		}

		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			log.Printf("INFO: Successfully read weather data from Tempest API - Station: %s", station.Name)
		}

		// Info level logging - show sensor data and night detection
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			isNight := isNightTime(obs.Illuminance)
			nightIndicator := ""
			if isNight {
				nightIndicator = " 🌙 NIGHT"
			}
			log.Printf("INFO: Sensor data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph (%.0f°), Rain: %.3f in, Light: %.0f lux%s",
				obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.WindDirection, obs.RainAccumulated, obs.Illuminance, nightIndicator)
		}

		// Debug logging - show all weather metrics and pretty printed JSON
		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: Full weather data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph (%.0f°), Rain: %.3f in, Pressure: %.1f mb, UV: %.1f, Solar: %.0f W/m², Battery: %.1fV",
				obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.WindDirection, obs.RainAccumulated,
				obs.StationPressure, obs.UV, obs.SolarRadiation, obs.Battery)

			// Pretty print the observation data as JSON
			jsonData, err := json.MarshalIndent(obs, "", "  ")
			if err == nil {
				log.Printf("DEBUG: Raw Tempest API JSON response:\n%s", string(jsonData))
			}

			log.Printf("DEBUG: Updating HomeKit accessories with new sensor values")
		}

		// Update HomeKit sensors with detailed logging
		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: HomeKit - Temperature: %.1f°C", obs.AirTemperature)
			log.Printf("DEBUG: HomeKit - Humidity: %.1f%%", obs.RelativeHumidity)
			log.Printf("DEBUG: HomeKit - Wind Speed: %.1f mph", obs.WindAvg)
			log.Printf("DEBUG: HomeKit - Wind Direction: %.0f°", obs.WindDirection)
			log.Printf("DEBUG: HomeKit - Rain Accumulation: %.3f in", obs.RainAccumulated)
			log.Printf("DEBUG: HomeKit - Illuminance: %.0f lux", obs.Illuminance)
		}

		wa.UpdateTemperature(obs.AirTemperature)
		wa.UpdateHumidity(obs.RelativeHumidity)
		wa.UpdateWindSpeed(obs.WindAvg)
		wa.UpdateWindDirection(obs.WindDirection)
		wa.UpdateRainAccumulation(obs.RainAccumulated)
		wa.UpdateIlluminance(obs.Illuminance)

		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: HomeKit accessory updates completed")
		}

		// Update web dashboard
		webServer.UpdateWeather(obs)

		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: Web dashboard updated with latest weather data")
		}
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

// isNightTime determines if it's nighttime based on illuminance levels
// Illuminance below 10 lux is generally considered nighttime
func isNightTime(illuminance float64) bool {
	return illuminance < 10.0
}
