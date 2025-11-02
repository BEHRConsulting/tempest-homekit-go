// Package weather provides a generated data source for synthetic weather data.
// This data source directly uses the weather generator without HTTP calls.
package weather

import (
	"time"

	"tempest-homekit-go/pkg/generator"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/types"
)

// GeneratedDataSource implements the DataSource interface for generated weather data.
// It directly uses a weather generator instead of making HTTP requests.
type GeneratedDataSource struct {
	stationID         int
	token             string
	stationName       string
	generator         generator.WeatherGenerator
	running           bool
	stopChan          chan struct{}
	observationChan   chan types.Observation
	latestObservation *types.Observation
	lastUpdate        time.Time
	observationCount  int64
}

// NewGeneratedDataSource creates a new generated data source
func NewGeneratedDataSource(stationID int, token, stationName string, gen generator.WeatherGenerator) *GeneratedDataSource {
	return &GeneratedDataSource{
		stationID:       stationID,
		token:           token,
		stationName:     stationName,
		generator:       gen,
		stopChan:        make(chan struct{}),
		observationChan: make(chan Observation, 10), // Buffer for observations
	}
}

// Start begins generating weather observations
func (g *GeneratedDataSource) Start() (<-chan types.Observation, error) {
	g.running = true

	// Start the generation loop in a goroutine
	go g.generationLoop()

	return g.observationChan, nil
}

// Stop halts the data source
func (g *GeneratedDataSource) Stop() error {
	if g.running {
		g.running = false
		close(g.stopChan)
	}
	return nil
}

// GetLatestObservation returns the most recent observation
func (g *GeneratedDataSource) GetLatestObservation() *Observation {
	return g.latestObservation
}

// GetForecast returns forecast data (generated data doesn't have forecasts)
func (g *GeneratedDataSource) GetForecast() *ForecastResponse {
	return nil
}

// GetStatus returns the current status
func (g *GeneratedDataSource) GetStatus() DataSourceStatus {
	location := g.generator.GetLocation()
	season := g.generator.GetSeason()

	return DataSourceStatus{
		Type:             DataSourceGenerated,
		Active:           g.running,
		LastUpdate:       g.lastUpdate,
		ObservationCount: g.observationCount,
		StationName:      g.stationName,
		Location:         location.Name,
		Season:           season.String(),
		ClimateZone:      location.ClimateZone,
	}
}

// GetType returns the data source type
func (g *GeneratedDataSource) GetType() DataSourceType {
	return DataSourceGenerated
}

// generationLoop generates weather observations every 60 seconds
func (g *GeneratedDataSource) generationLoop() {
	logger.Info("Starting generated data source generation loop (60 second interval)")

	// Generate initial observation
	g.generateObservation()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-g.stopChan:
			logger.Info("Generated data source generation loop stopped")
			return

		case <-ticker.C:
			g.generateObservation()
		}
	}
}

// generateObservation creates and sends a new weather observation
func (g *GeneratedDataSource) generateObservation() {
	logger.Debug("Generated data source: generating observation")

	// Generate a fresh observation
	obs := g.generator.GenerateObservation()
	if obs == nil {
		logger.Error("Failed to generate weather observation")
		return
	}

	// Update internal state
	g.latestObservation = obs
	g.lastUpdate = time.Now()
	g.observationCount++

	// Send to channel if running (non-blocking)
	if g.running {
		select {
		case g.observationChan <- *obs:
			logger.Debug("Generated observation sent to channel")
		default:
			logger.Debug("Generated observation channel full, skipping")
		}
	}
}
