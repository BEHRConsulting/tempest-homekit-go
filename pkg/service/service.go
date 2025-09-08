package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"tempest-homekit-go/pkg/config"
	// "tempest-homekit-go/pkg/homekit"
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
	// wa := homekit.NewWeatherAccessories()
	// t, err := homekit.SetupHomeKit(wa, cfg.Pin)
	// if err != nil {
	// 	return fmt.Errorf("failed to setup HomeKit: %v", err)
	// }

	// // Start HomeKit transport
	// go func() {
	// 	log.Println("Starting HomeKit transport...")
	// 	t.Start()
	// }()

	// Setup web dashboard
	webServer := web.NewWebServer(cfg.WebPort)
	log.Printf("Service: About to start web server goroutine")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WebServer Goroutine: Panic recovered: %v", r)
			}
		}()
		log.Printf("WebServer Goroutine: Starting...")
		log.Printf("Starting web dashboard on port %s", cfg.WebPort)
		if err := webServer.Start(); err != nil {
			log.Printf("Web server error: %v", err)
		}
		log.Printf("WebServer Goroutine: ListenAndServe returned, exiting goroutine")
	}()
	log.Printf("Service: Web server goroutine started")

	// Update HomeKit status in web server
	// homekitStatus := map[string]interface{}{
	// 	"bridge":      true,
	// 	"name":        "Tempest HomeKit Bridge",
	// 	"accessories": 6,
	// 	"accessoryNames": []string{
	// 		"Temperature Sensor",
	// 		"Humidity Sensor",
	// 		"Wind Speed Sensor",
	// 		"Wind Direction Sensor",
	// 		"Rain Sensor",
	// 		"Light Sensor",
	// 	},
	// 	"pin": cfg.Pin,
	// }
	// log.Printf("Sending HomeKit status to web server: %+v", homekitStatus)
	// webServer.UpdateHomeKitStatus(homekitStatus)
	// log.Printf("HomeKit status sent to web server successfully")

	// Poll weather data
	log.Println("Setting up weather polling...")
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	log.Printf("Ticker created with interval: %v", 60*time.Second)

	log.Println("Starting polling loop...")
	log.Printf("About to enter polling loop")

	for range ticker.C {
		log.Printf("Service: Ticker fired, starting polling iteration")
		obs, err := weather.GetObservation(station.StationID, cfg.Token)
		if err != nil {
			log.Printf("Error getting observation: %v", err)
			continue
		}

		log.Printf("Service: Got observation data - Temp: %.1f°C", obs.AirTemperature)

		// Info level logging - show sensor data
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			log.Printf("INFO: Sensor data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph (%.0f°), Rain: %.3f in",
				obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.WindDirection, obs.RainAccumulated)
		}

		// Debug logging - show all weather metrics and pretty printed JSON
		if cfg.LogLevel == "debug" {
			log.Printf("DEBUG: Full weather data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph (%.0f°), Rain: %.3f in, Pressure: %.1f mb, UV: %.1f, Solar: %.0f W/m², Battery: %.1fV",
				obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.WindDirection, obs.RainAccumulated,
				obs.StationPressure, obs.UV, obs.SolarRadiation, obs.Battery)

			// Pretty print the observation data as JSON
			jsonData, err := json.MarshalIndent(obs, "", "  ")
			if err == nil {
				log.Printf("DEBUG: Raw API data:\n%s", string(jsonData))
			}
		}

		log.Printf("Service: Updating HomeKit sensors")
		// wa.UpdateTemperature(obs.AirTemperature)
		// wa.UpdateHumidity(obs.RelativeHumidity)
		// wa.UpdateWindSpeed(obs.WindAvg)
		// wa.UpdateWindDirection(obs.WindDirection)
		// wa.UpdateRainAccumulation(obs.RainAccumulated)
		// wa.UpdateIlluminance(obs.Illuminance)

		// Update web dashboard
		log.Printf("Service: About to call webServer.UpdateWeather")
		webServer.UpdateWeather(obs)
		log.Printf("Service: webServer.UpdateWeather call completed")
		log.Printf("Service: Polling loop iteration completed")
	}
	log.Printf("Service: StartService function is returning")
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
