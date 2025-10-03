// Package weather defines the unified data source interface for weather observations.
// All weather data sources (API, UDP, Generated) implement this interface.
package weather

import (
	"time"
)

// DataSource is the unified interface for all weather data sources.
// Implementations include: API polling, UDP stream, and generated weather.
type DataSource interface {
	// Start begins the data source and returns a channel for observations
	Start() (<-chan Observation, error)

	// Stop gracefully shuts down the data source
	Stop() error

	// GetLatestObservation returns the most recent observation
	GetLatestObservation() *Observation

	// GetForecast returns forecast data (may return nil if not available)
	GetForecast() *ForecastResponse

	// GetStatus returns data source status information
	GetStatus() DataSourceStatus

	// GetType returns the type of data source
	GetType() DataSourceType
}

// DataSourceType identifies the type of weather data source
type DataSourceType string

const (
	// DataSourceAPI represents WeatherFlow API polling
	DataSourceAPI DataSourceType = "api"

	// DataSourceUDP represents local UDP broadcast listener
	DataSourceUDP DataSourceType = "udp"

	// DataSourceGenerated represents simulated weather data
	DataSourceGenerated DataSourceType = "generated"

	// DataSourceCustomURL represents a custom station URL
	DataSourceCustomURL DataSourceType = "custom-url"
)

// DataSourceStatus provides unified status information for any data source
type DataSourceStatus struct {
	Type             DataSourceType `json:"type"`
	Active           bool           `json:"active"`
	LastUpdate       time.Time      `json:"lastUpdate"`
	ObservationCount int64          `json:"observationCount"`

	// Optional fields depending on source type
	StationName  string `json:"stationName,omitempty"`
	StationIP    string `json:"stationIP,omitempty"`    // For UDP
	SerialNumber string `json:"serialNumber,omitempty"` // For UDP
	PacketCount  int64  `json:"packetCount,omitempty"`  // For UDP
	Location     string `json:"location,omitempty"`     // For Generated
	Season       string `json:"season,omitempty"`       // For Generated
	ClimateZone  string `json:"climateZone,omitempty"`  // For Generated
	CustomURL    string `json:"customURL,omitempty"`    // For Custom URL
}
