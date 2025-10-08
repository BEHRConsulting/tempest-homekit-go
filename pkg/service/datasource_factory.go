// Package service provides data source factory for creating appropriate weather data sources.
package service

import (
	"fmt"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
)

// CreateDataSource examines config flags and returns the appropriate DataSource implementation.
// This is the only place in the codebase that needs to know about different data source types.
// For UDP mode, the udpListener parameter must be provided as interface{} (pass nil for other modes).
// This uses interface{} to avoid import cycle between weather and udp packages.
func CreateDataSource(cfg *config.Config, station *weather.Station, udpListener interface{}) (weather.DataSource, error) {
	// Priority order:
	// 1. UDP Stream (if enabled)
	// 2. Custom Station URL (if provided)
	// 3. Generated Weather (if enabled)
	// 4. WeatherFlow API (default)

	if cfg.UDPStream {
		if udpListener == nil {
			return nil, fmt.Errorf("UDP listener required for UDP stream mode")
		}

		logger.Info("Creating UDP data source (offline mode: %v)", cfg.DisableInternet)

		// Create UDP data source wrapper
		var stationID int
		var token string
		if station != nil {
			stationID = station.StationID
		}
		if !cfg.DisableInternet {
			token = cfg.Token
		}

		// Type assert to the interface defined in weather package
		listener, ok := udpListener.(weather.UDPListener)
		if !ok {
			return nil, fmt.Errorf("invalid UDP listener type")
		}

		dataSource := weather.NewUDPDataSource(listener, cfg.DisableInternet, stationID, token)
		logger.Info("✓ UDP data source created (port 50222)")
		return dataSource, nil
	}

	if cfg.StationURL != "" {
		logger.Info("Creating API data source with custom URL: %s", cfg.StationURL)

		// Custom station URL (generated weather, etc.)
		var stationID int
		var stationName string
		if station != nil {
			stationID = station.StationID
			stationName = station.StationName
		}

		dataSource := weather.NewAPIDataSource(stationID, cfg.Token, stationName, weather.APIDataSourceOptions{CustomURL: cfg.StationURL, GeneratedPath: cfg.GeneratedWeatherPath})
		logger.Info("✓ API data source created with custom URL")
		return dataSource, nil
	}

	if cfg.UseGeneratedWeather {
		logger.Info("Creating API data source with generated weather")

		// Generated weather uses the internal endpoint
		var stationID int
		var stationName string
		if station != nil {
			stationID = station.StationID
			stationName = station.StationName
		}

		// Ensure we construct the generated URL using the configured path.
		// Tests sometimes construct a Config manually and may leave WebPort or
		// GeneratedWeatherPath empty; default to historical defaults so behavior
		// remains predictable in those cases.
		port := cfg.WebPort
		if port == "" {
			port = "8080"
		}
		path := cfg.GeneratedWeatherPath
		if path == "" {
			path = "/api/generate-weather"
		}
		generatedURL := fmt.Sprintf("http://localhost:%s%s", port, path)
		dataSource := weather.NewAPIDataSource(stationID, cfg.Token, stationName, weather.APIDataSourceOptions{CustomURL: generatedURL, GeneratedPath: cfg.GeneratedWeatherPath})
		logger.Info("✓ Generated weather data source created")
		return dataSource, nil
	}

	// Default: WeatherFlow API
	if station == nil {
		return nil, fmt.Errorf("station required for API data source")
	}

	logger.Info("Creating API data source for station: %s (ID: %d)", station.StationName, station.StationID)
	dataSource := weather.NewAPIDataSource(station.StationID, cfg.Token, station.StationName, weather.APIDataSourceOptions{CustomURL: "", GeneratedPath: cfg.GeneratedWeatherPath})
	logger.Info("✓ WeatherFlow API data source created")
	return dataSource, nil
}
