// Package service provides the main service orchestration for the Tempest HomeKit bridge.
// It coordinates between the WeatherFlow API client, HomeKit accessories, and web dashboard.
package service

import (
	"fmt"
	"time"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/generator"
	"tempest-homekit-go/pkg/homekit"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/udp"
	"tempest-homekit-go/pkg/weather"
	"tempest-homekit-go/pkg/web"
)

// StartService initializes and starts the Tempest HomeKit service with the provided configuration.
// It sets up HomeKit accessories, starts the web server, and begins weather data polling.
// Now uses unified data source architecture for clean separation of concerns.
func StartService(cfg *config.Config, version string) error {
	// Set log level
	logger.SetLogLevel(cfg.LogLevel)

	logger.Info("Starting Tempest HomeKit service...")

	// Step 1: Get station information based on mode
	var station *weather.Station
	var weatherGen *generator.WeatherGenerator

	if cfg.UDPStream {
		// UDP mode - create placeholder station
		logger.Info("UDP stream mode - will create UDP data source later")
		station = &weather.Station{
			StationID:   0,
			Name:        cfg.StationName,
			StationName: cfg.StationName,
		}
		if cfg.DisableInternet {
			logger.Info("Running in offline mode (--disable-internet) - all internet access disabled")
		}
	} else if cfg.UseGeneratedWeather {
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

	// Create web server only if not disabled
	var webServer *web.WebServer
	if !cfg.DisableWebConsole {
		webServer = web.NewWebServer(cfg.WebPort, cfg.Elevation, cfg.LogLevel, station.StationID, cfg.UseWebStatus, version, effectiveStationURL, generatedWeatherInfo, weatherGen, cfg.Units, cfg.UnitsPressure)
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
	} else {
		logger.Info("Web console disabled (--disable-webconsole)")
	}

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
	if webServer != nil {
		webServer.UpdateHomeKitStatus(homekitStatus)
	}

	// Preload historical data if requested
	if cfg.ReadHistory {
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			logger.Info("--read-history flag detected, preloading last 24 hours of weather data...")
		}

		// Create a progress callback function
		progressCallback := func(currentStep, totalSteps int, description string) {
			if webServer != nil {
				webServer.SetHistoryLoadingProgress(currentStep, totalSteps, description)
			}
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
				if webServer != nil {
					webServer.SetHistoryLoadingComplete()
				}
			} else {
				logger.Debug("Successfully fetched %d historical observations", len(historicalObs))
			}
		}

		if err == nil && webServer != nil {
			webServer.SetHistoryLoadingProgress(2, 3, "Processing historical data...")

			// Send historical data to web server for charts
			for _, obs := range historicalObs {
				webServer.UpdateWeather(obs)
				logger.Debug("Added historical observation from %v", time.Unix(obs.Timestamp, 0))
			}

			// Complete the loading process
			webServer.SetHistoryLoadingComplete()

			webServer.SetHistoricalDataStatus(len(historicalObs))

			if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
				logger.Info("Historical data preload completed - loaded %d observations", len(historicalObs))
			}
		}
	}

	// UNIFIED DATA SOURCE APPROACH
	// Create UDP listener if needed (service layer handles this to avoid import cycles)
	var udpListener *udp.UDPListener
	if cfg.UDPStream {
		logger.Info("Creating UDP listener for UDP stream mode")
		udpListener = udp.NewUDPListener()
	}

	// Create appropriate data source using factory pattern
	logger.Info("Creating data source...")
	dataSource, err := CreateDataSource(cfg, station, udpListener)
	if err != nil {
		return fmt.Errorf("failed to create data source: %v", err)
	}
	defer dataSource.Stop()

	// Start the data source
	logger.Info("Starting data source: %s", dataSource.GetType())
	obsChan, err := dataSource.Start()
	if err != nil {
		return fmt.Errorf("failed to start data source: %v", err)
	}

	// Set initial data source status in web server (before any observations arrive)
	if webServer != nil {
		initialStatus := dataSource.GetStatus()
		webServer.UpdateDataSourceStatus(initialStatus)
		logger.Debug("Initial data source status set: type=%s", initialStatus.Type)
	}

	// Main observation processing loop - unified for all data sources!
	logger.Info("Starting unified observation processing loop")
	for obs := range obsChan {
		logger.Debug("Processing observation from %s data source", dataSource.GetType())

		// Update HomeKit sensors (if enabled)
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
			logger.Debug("HomeKit sensors updated")
		}

		// Update web server
		if webServer != nil {
			webServer.UpdateWeather(&obs)
			logger.Debug("Web server updated")

			// Update forecast from data source (if available)
			if forecast := dataSource.GetForecast(); forecast != nil {
				webServer.UpdateForecast(forecast)
				logger.Debug("Forecast updated")
			}

			// Update data source status in web server
			status := dataSource.GetStatus()
			webServer.UpdateDataSourceStatus(status)
			logger.Debug("Data source status updated")
		}

		// Log observation details
		if cfg.LogLevel == "info" || cfg.LogLevel == "debug" {
			logger.Info("%s data - Temp: %.1f°C, Humidity: %.1f%%, Wind: %.1f m/s, Lux: %.0f",
				dataSource.GetType(), obs.AirTemperature, obs.RelativeHumidity, obs.WindAvg, obs.Illuminance)
		}
	}

	logger.Info("Observation processing loop ended")
	return nil
}
