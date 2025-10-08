// Package weather provides the API data source implementation.
package weather

import (
	"net/url"
	"sync"
	"time"

	"tempest-homekit-go/pkg/logger"
)

// APIDataSource implements DataSource for WeatherFlow API polling
type APIDataSource struct {
	stationID     int
	token         string
	stationName   string
	customURL     string // For custom station URLs (generated weather, etc.)
	generated     bool   // true when this API data source is using generated weather endpoint
	generatedPath string // configured path used to identify generated weather endpoint

	mu                sync.RWMutex
	latestObservation *Observation
	latestForecast    *ForecastResponse
	observationChan   chan Observation
	stopChan          chan struct{}
	observationCount  int64
	lastUpdate        time.Time
	running           bool
}

// APIDataSourceOptions holds optional parameters for creating APIDataSource
type APIDataSourceOptions struct {
	CustomURL     string
	GeneratedPath string
}

// NewAPIDataSource creates a new API-based data source with options.
func NewAPIDataSource(stationID int, token, stationName string, opts APIDataSourceOptions) *APIDataSource {
	a := &APIDataSource{
		stationID:       stationID,
		token:           token,
		stationName:     stationName,
		customURL:       opts.CustomURL,
		generatedPath:   opts.GeneratedPath,
		observationChan: make(chan Observation, 100),
		stopChan:        make(chan struct{}),
	}

	// Default generatedPath when empty
	if a.generatedPath == "" {
		a.generatedPath = "/api/generate-weather"
	}

	// Determine if this data source points to the generated weather endpoint by
	// parsing the URL and comparing the path exactly to the configured generatedPath.
	if a.customURL != "" {
		if u, err := url.Parse(a.customURL); err == nil {
			if u.Path == a.generatedPath {
				a.generated = true
			}
		}
	}

	return a
}

// Start begins polling the API for weather data
func (a *APIDataSource) Start() (<-chan Observation, error) {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return a.observationChan, nil
	}
	a.running = true
	a.mu.Unlock()

	// Start polling goroutine
	go a.pollLoop()

	return a.observationChan, nil
}

// Stop gracefully shuts down the API data source
func (a *APIDataSource) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	close(a.stopChan)
	a.running = false
	close(a.observationChan)

	logger.Info("API data source stopped")
	return nil
}

// GetLatestObservation returns the most recent observation
func (a *APIDataSource) GetLatestObservation() *Observation {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.latestObservation
}

// GetForecast returns the latest forecast data
func (a *APIDataSource) GetForecast() *ForecastResponse {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.latestForecast
}

// GetStatus returns the current status of the API data source
func (a *APIDataSource) GetStatus() DataSourceStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sourceType := DataSourceAPI
	if a.customURL != "" {
		if a.generated {
			sourceType = DataSourceGenerated
		} else {
			sourceType = DataSourceCustomURL
		}
	}

	return DataSourceStatus{
		Type:             sourceType,
		Active:           a.running,
		LastUpdate:       a.lastUpdate,
		ObservationCount: a.observationCount,
		StationName:      a.stationName,
		CustomURL:        a.customURL,
	}
}

// GetType returns the data source type
func (a *APIDataSource) GetType() DataSourceType {
	if a.customURL != "" {
		if a.generated {
			return DataSourceGenerated
		}
		return DataSourceCustomURL
	}
	return DataSourceAPI
}

// pollLoop is the main polling loop that fetches data every 60 seconds
func (a *APIDataSource) pollLoop() {
	logger.Info("Starting API data source polling loop (60 second interval)")

	// Initial fetch
	a.fetchObservation()
	a.fetchForecast()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	forecastCounter := 0

	for {
		select {
		case <-a.stopChan:
			logger.Info("API polling loop stopped")
			return

		case <-ticker.C:
			a.fetchObservation()

			// Update forecast every 30 minutes (30 ticks)
			forecastCounter++
			if forecastCounter >= 30 {
				a.fetchForecast()
				forecastCounter = 0
			}
		}
	}
}

// fetchObservation retrieves a single observation from the API
func (a *APIDataSource) fetchObservation() {
	logger.Debug("API data source: fetching observation")

	var obs *Observation
	var err error

	if a.customURL != "" {
		// Use custom station URL (generated weather, etc.)
		obs, err = GetObservationFromURL(a.customURL)
		if err != nil {
			logger.Error("Error getting observation from URL %s: %v", a.customURL, err)
			return
		}
		logger.Debug("Successfully fetched observation from custom URL: %s", a.customURL)
	} else {
		// Use real Tempest API
		obs, err = GetObservation(a.stationID, a.token)
		if err != nil {
			logger.Error("Error getting observation from API: %v", err)
			return
		}
		logger.Debug("Successfully fetched observation from WeatherFlow API")
	}

	if obs != nil {
		a.mu.Lock()
		a.latestObservation = obs
		a.lastUpdate = time.Now()
		a.observationCount++
		a.mu.Unlock()

		// Send to channel (non-blocking)
		select {
		case a.observationChan <- *obs:
			logger.Debug("Observation sent to channel")
		default:
			logger.Debug("Observation channel full, skipping")
		}
	}
}

// fetchForecast retrieves forecast data from the API
func (a *APIDataSource) fetchForecast() {
	// Skip forecast for generated weather
	if a.generated {
		logger.Debug("Skipping forecast fetch for generated weather")
		return
	}

	logger.Debug("API data source: fetching forecast")

	forecast, err := GetForecast(a.stationID, a.token)
	if err != nil {
		logger.Error("Error getting forecast: %v", err)
		return
	}

	a.mu.Lock()
	a.latestForecast = forecast
	a.mu.Unlock()

	logger.Debug("Successfully fetched forecast data")
}
