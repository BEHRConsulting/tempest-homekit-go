// Package service provides the main service orchestration for the Tempest HomeKit bridge.
// It coordinates between the WeatherFlow API client, HomeKit accessories, and web dashboard.
package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/generator"
	"tempest-homekit-go/pkg/homekit"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
	"tempest-homekit-go/pkg/web"
)

// StartService initializes and starts the Tempest HomeKit service with the provided configuration.
// It sets up HomeKit accessories, starts the web server, and begins weather data polling.
func StartService(cfg *config.Config, version string) error {
	// Set log level
	logger.SetLogLevel(cfg.LogLevel)

	logger.Info("Starting Tempest HomeKit service...")

	var station *weather.Station
	var weatherGen *generator.WeatherGenerator

	if cfg.UseGeneratedWeather {
		// Use generated weather data for testing
		logger.Info("Using generated weather data for testing")
		weatherGen = generator.NewWeatherGenerator()

		// Create a fake station for the generated location
		location := weatherGen.GetLocation()
		station = &weather.Station{
			StationID:   99999, // Fake station ID
			Name:        location.Name,
			StationName: location.Name,
		}

		// Update the config elevation to match the generated location
		cfg.Elevation = location.Elevation
		logger.Info("Using generated location elevation: %.1f meters (%.0f feet)", location.Elevation, location.Elevation*3.28084)

		logger.Info("Generated weather location: %s (%s, %s season)",
			location.Name, location.ClimateZone, weatherGen.GetSeason().String())
	} else {
		// Use real Tempest API data
		logger.Debug("Fetching stations from WeatherFlow API")
		stations, err := weather.GetStations(cfg.Token)
		if err != nil {
			return fmt.Errorf("failed to get stations: %v", err)
		}

		station = weather.FindStationByName(stations, cfg.StationName)
		if station == nil {
			logger.Info("Available stations:")
			for _, s := range stations {
				logger.Info("  - ID: %d, Name: '%s', StationName: '%s'", s.StationID, s.Name, s.StationName)
			}
			return fmt.Errorf("station '%s' not found", cfg.StationName)
		}

		logger.Info("Found station: %s (ID: %d)", station.Name, station.StationID)
	}

	// Parse sensor configuration (needed for both HomeKit and web server)
	sensorConfig := config.ParseSensorConfig(cfg.Sensors)

	// Conditionally setup HomeKit based on configuration
	var ws *homekit.WeatherSystemModern
	if cfg.DisableHomeKit {
		logger.Info("HomeKit services disabled - running in web console only mode")
	} else {
		// Setup HomeKit with sensor configuration
		logger.Debug("Initializing HomeKit accessories with sensor config: %s", cfg.Sensors)
		var setupErr error
		ws, setupErr = homekit.NewWeatherSystemModern(cfg.Pin, &sensorConfig, cfg.LogLevel)
		if setupErr != nil {
			return fmt.Errorf("failed to setup HomeKit: %v", setupErr)
		}

		// Start the HomeKit server
		logger.Debug("Starting weather system server")
		go func() {
			if err := ws.Start(); err != nil {
				logger.Error("HomeKit server error: %v", err)
			}
		}()

		logger.Info("HomeKit server started successfully with PIN: %s", cfg.Pin)
		logger.Debug("HomeKit - Bridge ready to accept connections")
		logger.Debug("HomeKit - Listening for iOS/HomeKit client connections...")
	}

	// Setup web dashboard
	var generatedWeatherInfo *web.GeneratedWeatherInfo
	if cfg.UseGeneratedWeather {
		location := weatherGen.GetLocation()
		generatedWeatherInfo = &web.GeneratedWeatherInfo{
			Enabled:     true,
			Location:    location.Name,
			Season:      weatherGen.GetSeason().String(),
			ClimateZone: location.ClimateZone,
		}
	}

	// Determine the effective station URL that will be used for weather data
	effectiveStationURL := cfg.StationURL
	if effectiveStationURL == "" {
		// Construct the actual Tempest API URL that will be used
		effectiveStationURL = fmt.Sprintf("https://swd.weatherflow.com/swd/rest/observations/station/%d?token=%s", station.StationID, cfg.Token)
	}

	webServer := web.NewWebServer(cfg.WebPort, cfg.Elevation, cfg.LogLevel, station.StationID, cfg.UseWebStatus, version, effectiveStationURL, generatedWeatherInfo, weatherGen, cfg.Units, cfg.UnitsPressure)
	webServer.SetStationName(station.Name)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Web server panic recovered: %v", r)
			}
		}()
		logger.Info("Starting web dashboard on port %s", cfg.WebPort)
		if err := webServer.Start(); err != nil {
			logger.Error("Web server error: %v", err)
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

	// Update HomeKit status in web server based on whether HomeKit is enabled
	var homekitStatus map[string]interface{}
	if cfg.DisableHomeKit {
		homekitStatus = map[string]interface{}{
			"bridge":         false,
			"name":           "HomeKit Disabled",
			"accessories":    0,
			"accessoryNames": []string{},
			"sensorConfig":   "Web Console Only",
			"pin":            "N/A",
			"status":         "Disabled by --disable-homekit flag",
		}
	} else {
		homekitStatus = map[string]interface{}{
			"bridge":         true,
			"name":           "Tempest HomeKit Bridge",
			"accessories":    len(enabledSensors),
			"accessoryNames": enabledSensors,
			"sensorConfig":   cfg.Sensors,
			"pin":            cfg.Pin,
		}
	}
	webServer.UpdateHomeKitStatus(homekitStatus)

	// Preload historical data if requested
	if cfg.ReadHistory {
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			logger.Info("--read-history flag detected, preloading last 24 hours of weather data...")
		}

		// Create a progress callback function
		progressCallback := func(currentStep, totalSteps int, description string) {
			webServer.SetHistoryLoadingProgress(currentStep, totalSteps, description)
		}

		var historicalObs []*weather.Observation
		var err error

		if cfg.UseGeneratedWeather && weatherGen != nil {
			// Generate historical data
			logger.Info("Generating 1000 historical weather data points...")
			historicalObs = weatherGen.GenerateHistoricalData(1000)
			logger.Debug("Successfully generated %d historical observations", len(historicalObs))
		} else {
			// Use real historical data from API
			historicalObs, err = weather.GetHistoricalObservationsWithProgress(station.StationID, cfg.Token, cfg.LogLevel, progressCallback)
			if err != nil {
				logger.Error("Failed to fetch historical data: %v", err)
				webServer.SetHistoryLoadingComplete()
			} else {
				logger.Debug("Successfully fetched %d historical observations", len(historicalObs))
			}
		}

		if err == nil {
			webServer.SetHistoryLoadingProgress(2, 3, "Processing historical data...")

			// Send historical data to web server for charts
			for _, obs := range historicalObs {
				webServer.UpdateWeather(obs)
				logger.Debug("Added historical observation from %v", time.Unix(obs.Timestamp, 0))
			}

			// Complete the loading process
			webServer.SetHistoryLoadingComplete()

			// Update historical data status in web server
			webServer.SetHistoricalDataStatus(len(historicalObs))

			if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
				logger.Info("Historical data preload completed - loaded %d observations", len(historicalObs))
			}
		}
	}

	// Poll weather data
	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		logger.Info("Setting up weather polling every 60 seconds")
	}
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial data fetch to populate HomeKit immediately
	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		logger.Info("Fetching initial weather data to populate HomeKit")
	}
	updateWeatherData(station, cfg, ws, webServer, weatherGen)

	// Fetch initial forecast data (skip for generated weather)
	if !cfg.UseGeneratedWeather {
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			logger.Info("Fetching initial forecast data")
		}
		updateForecastData(station, cfg, webServer)
	}

	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		logger.Info("Starting weather data polling loop")
	}

	forecastUpdateCounter := 0
	for range ticker.C {
		updateWeatherData(station, cfg, ws, webServer, weatherGen)

		// Update forecast every 30 minutes (30 ticks) - skip for generated weather
		if !cfg.UseGeneratedWeather {
			forecastUpdateCounter++
			if forecastUpdateCounter >= 30 {
				updateForecastData(station, cfg, webServer)
				forecastUpdateCounter = 0
			}
		}
	}
	return nil
}

func updateWeatherData(station *weather.Station, cfg *config.Config, ws *homekit.WeatherSystemModern, webServer *web.WebServer, weatherGen *generator.WeatherGenerator) {
	logger.Debug("Polling iteration started - fetching observation from station %d", station.StationID)

	var obs *weather.Observation
	var err error

	if cfg.StationURL != "" {
		// Use custom station URL (e.g., generated weather endpoint)
		obs, err = weather.GetObservationFromURL(cfg.StationURL)
		if err != nil {
			logger.Error("Error getting observation from URL %s: %v", cfg.StationURL, err)
			return
		}
		logger.Info("Successfully read weather data from custom station URL: %s", cfg.StationURL)
	} else {
		// Use real Tempest API data
		obs, err = weather.GetObservation(station.StationID, cfg.Token)
		if err != nil {
			logger.Error("Error getting observation: %v", err)
			return
		}
		logger.Info("Successfully read weather data from Tempest API - Station: %s", station.Name)
	}

	// Info level logging - show sensor data and night detection
	if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
		isNight := isNightTime(obs.Illuminance)
		nightIndicator := ""
		if isNight {
			nightIndicator = " 🌙 NIGHT"
		}
		logger.Info("Sensor data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph (%.0f°), Rain: %.3f in, Light: %.0f lux%s",
			obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.WindDirection, obs.RainAccumulated, obs.Illuminance, nightIndicator)
	}

	// Debug logging - show all weather metrics and pretty printed JSON
	logger.Debug("Full weather data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f mph (%.0f°), Rain: %.3f in, Pressure: %.1f mb, UV: %d, Solar: %.0f W/m², Battery: %.1fV",
		obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.WindDirection, obs.RainAccumulated,
		obs.StationPressure, obs.UV, obs.SolarRadiation, obs.Battery)

	// Pretty print the observation data as JSON
	jsonData, err := json.MarshalIndent(obs, "", "  ")
	if err == nil {
		logger.Debug("Raw Tempest API JSON response:\n%s", string(jsonData))
	}

	logger.Debug("Updating HomeKit accessories with new sensor values")

	// Update HomeKit sensors with detailed logging
	if cfg.LogLevel == "debug" {
		logger.Debug("HomeKit - Air Temperature: %.1f°C", obs.AirTemperature)
		logger.Debug("HomeKit - Relative Humidity: %.1f%%", obs.RelativeHumidity)
		logger.Debug("HomeKit - Wind Average: %.1f mph", obs.WindAvg)
		logger.Debug("HomeKit - Wind Gust: %.1f mph", obs.WindGust)
		logger.Debug("HomeKit - Wind Direction: %.0f°", obs.WindDirection)
		logger.Debug("HomeKit - Rain: %.3f in", obs.RainAccumulated)
		logger.Debug("HomeKit - Precipitation Type: %d", obs.PrecipitationType)
		logger.Debug("HomeKit - Lightning Distance: %.0f", obs.LightningStrikeAvg)
		logger.Debug("HomeKit - Lightning Count: %d", obs.LightningStrikeCount)
		logger.Debug("HomeKit - Lux: %.0f lux", obs.Illuminance)
		logger.Debug("HomeKit - UV Index: %d", obs.UV)
	}

	// Update all sensors in the Tempest Weather Station using the new UpdateSensor method:
	if ws != nil {
		ws.UpdateSensor("Wind Speed", obs.WindAvg)
		ws.UpdateSensor("Wind Gust", obs.WindGust)
		ws.UpdateSensor("Wind Direction", obs.WindDirection)
		ws.UpdateSensor("Air Temperature", obs.AirTemperature)
		ws.UpdateSensor("Relative Humidity", obs.RelativeHumidity)
		ws.UpdateSensor("Ambient Light", obs.Illuminance)
		ws.UpdateSensor("UV Index", float64(obs.UV))
		ws.UpdateSensor("Rain Accumulation", obs.RainAccumulated)
		ws.UpdateSensor("Precipitation Type", float64(obs.PrecipitationType))
		ws.UpdateSensor("Lightning Count", float64(obs.LightningStrikeCount))
		ws.UpdateSensor("Lightning Distance", obs.LightningStrikeAvg)
	} else {
		logger.Debug("Skipping HomeKit sensor updates - HomeKit disabled")
	}

	if cfg.LogLevel == "debug" {
		logger.Debug("HomeKit accessory updates completed - ULTRA-MINIMAL: Only temperature sensor active")
		logger.Debug("HomeKit - All other sensors ignored for maximum compliance")
		logger.Debug("HomeKit - Temperature characteristic changes pushed to connected iOS devices")
	}

	// Update web dashboard
	webServer.UpdateWeather(obs)

	if cfg.LogLevel == "debug" {
		logger.Debug("Web dashboard updated with latest weather data")
	}

	// Update battery data in status manager if using fallback status
	if !cfg.UseWebStatus {
		webServer.UpdateBatteryFromObservation(obs)
	}
}

func updateForecastData(station *weather.Station, cfg *config.Config, webServer *web.WebServer) {
	logger.Debug("Fetching forecast data for station %d", station.StationID)

	forecast, err := weather.GetForecast(station.StationID, cfg.Token)
	if err != nil {
		log.Printf("WARNING: Failed to fetch forecast data: %v", err)
		return
	}

	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Successfully fetched forecast data with %d daily periods", len(forecast.Forecast.Daily))
	}

	// Update web server with forecast data
	webServer.UpdateForecast(forecast)

	if cfg.LogLevel == "debug" {
		log.Printf("DEBUG: Web dashboard updated with latest forecast data")
	}
}

// isNightTime determines if it's nighttime based on illuminance levels
// Illuminance below 10 lux is generally considered nighttime
func isNightTime(illuminance float64) bool {
	return illuminance < 10.0
}
