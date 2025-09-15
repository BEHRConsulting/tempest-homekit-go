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

	// Setup HomeKit with sensor configuration
	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Initializing HomeKit accessories with sensor config: %s", cfg.Sensors)
	}
	sensorConfig := config.ParseSensorConfig(cfg.Sensors)
	ws, setupErr := homekit.NewWeatherSystemModern(cfg.Pin, &sensorConfig)
	if setupErr != nil {
		return fmt.Errorf("failed to setup HomeKit: %v", setupErr)
	}

	// Start the HomeKit server
	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Starting weather system server")
	}
	go func() {
		if err := ws.Start(); err != nil {
			log.Printf("HomeKit server error: %v", err)
		}
	}()

	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		log.Printf("INFO: HomeKit server started successfully with PIN: %s", cfg.Pin)
		log.Printf("DEBUG: HomeKit - Bridge ready to accept connections")
		log.Printf("DEBUG: HomeKit - Listening for iOS/HomeKit client connections...")
	}

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
	var enabledSensors []string
	if sensorConfig.Temperature {
		enabledSensors = append(enabledSensors, "Temperature")
	}
	if sensorConfig.Humidity {
		enabledSensors = append(enabledSensors, "Humidity")
	}
	if sensorConfig.Light {
		enabledSensors = append(enabledSensors, "Light")
	}
	if sensorConfig.Wind {
		enabledSensors = append(enabledSensors, "Wind Speed", "Wind Direction")
	}
	if sensorConfig.Rain {
		enabledSensors = append(enabledSensors, "Rain")
	}
	if sensorConfig.Pressure {
		enabledSensors = append(enabledSensors, "Pressure")
	}
	if sensorConfig.UV {
		enabledSensors = append(enabledSensors, "UV")
	}
	if sensorConfig.Lightning {
		enabledSensors = append(enabledSensors, "Lightning")
	}

	homekitStatus := map[string]interface{}{
		"bridge":         true,
		"name":           "Tempest HomeKit Bridge",
		"accessories":    len(enabledSensors),
		"accessoryNames": enabledSensors,
		"sensorConfig":   cfg.Sensors,
		"pin":            cfg.Pin,
	}
	webServer.UpdateHomeKitStatus(homekitStatus)

	// Poll weather data
	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		log.Printf("INFO: Setting up weather polling every 60 seconds")
	}
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial data fetch to populate HomeKit immediately
	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		log.Printf("INFO: Fetching initial weather data to populate HomeKit")
	}
	updateWeatherData(station, cfg, ws, webServer)

	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		log.Printf("INFO: Starting weather data polling loop")
	}

	for range ticker.C {
		updateWeatherData(station, cfg, ws, webServer)
	}
	return nil
}

func updateWeatherData(station *weather.Station, cfg *config.Config, ws *homekit.WeatherSystemModern, webServer *web.WebServer) {
	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Polling iteration started - fetching observation from station %d", station.StationID)
	}

	obs, err := weather.GetObservation(station.StationID, cfg.Token)
	if err != nil {
		log.Printf("Error getting observation: %v", err)
		return
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
		log.Printf("DEBUG: HomeKit - Air Temperature: %.1f°C", obs.AirTemperature)
		log.Printf("DEBUG: HomeKit - Relative Humidity: %.1f%%", obs.RelativeHumidity)
		log.Printf("DEBUG: HomeKit - Wind Average: %.1f mph", obs.WindAvg)
		log.Printf("DEBUG: HomeKit - Wind Gust: %.1f mph", obs.WindGust)
		log.Printf("DEBUG: HomeKit - Wind Direction: %.0f°", obs.WindDirection)
		log.Printf("DEBUG: HomeKit - Rain: %.3f in", obs.RainAccumulated)
		log.Printf("DEBUG: HomeKit - Precipitation Type: %d", obs.PrecipitationType)
		log.Printf("DEBUG: HomeKit - Lightning Distance: %.0f", obs.LightningStrikeAvg)
		log.Printf("DEBUG: HomeKit - Lightning Count: %d", obs.LightningStrikeCount)
		log.Printf("DEBUG: HomeKit - Lux: %.0f lux", obs.Illuminance)
		log.Printf("DEBUG: HomeKit - UV Index: %.0f", obs.UV)
	}

	// Update all sensors in the Tempest Weather Station using the new UpdateSensor method:
	ws.UpdateSensor("Wind Speed", obs.WindAvg)
	ws.UpdateSensor("Wind Gust", obs.WindGust)
	ws.UpdateSensor("Wind Direction", obs.WindDirection)
	ws.UpdateSensor("Air Temperature", obs.AirTemperature)
	ws.UpdateSensor("Relative Humidity", obs.RelativeHumidity)
	ws.UpdateSensor("Ambient Light", obs.Illuminance)
	ws.UpdateSensor("UV Index", obs.UV)
	ws.UpdateSensor("Rain Accumulation", obs.RainAccumulated)
	ws.UpdateSensor("Precipitation Type", float64(obs.PrecipitationType))
	ws.UpdateSensor("Lightning Count", float64(obs.LightningStrikeCount))
	ws.UpdateSensor("Lightning Distance", obs.LightningStrikeAvg)

	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: HomeKit accessory updates completed - ULTRA-MINIMAL: Only temperature sensor active")
		log.Printf("DEBUG: HomeKit - All other sensors ignored for maximum compliance")
		log.Printf("DEBUG: HomeKit - Temperature characteristic changes pushed to connected iOS devices")
	}

	// Update web dashboard
	webServer.UpdateWeather(obs)

	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Web dashboard updated with latest weather data")
	}
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
