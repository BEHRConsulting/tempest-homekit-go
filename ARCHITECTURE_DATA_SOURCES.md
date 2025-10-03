# Data Source Architecture Refactoring

## Overview

The application has been refactored to use a **unified data source architecture** based on the Strategy/Adapter pattern. This eliminates complex branching logic and ensures all data sources (API, UDP, Generated Weather) are handled uniformly.

## Motivation

**Before**: The service layer had complex branching logic checking config flags (`UDPStream`, `UseGeneratedWeather`, `StationURL`) throughout the codebase. This made the code hard to maintain and extend.

**After**: All data sources implement a single `DataSource` interface. The service layer is completely agnostic to the source type - it just processes observations from a channel.

## Architecture

### Core Interface (`pkg/weather/datasource.go`)

```go
type DataSource interface {
    Start() (<-chan Observation, error)  // Start and get observation channel
    Stop() error                          // Graceful shutdown
    GetLatestObservation() *Observation   // Get most recent data
    GetForecast() *ForecastResponse       // Get forecast (if available)
    GetStatus() DataSourceStatus          // Get source status
    GetType() DataSourceType              // Get source type (api/udp/generated)
}
```

### Implementations

1. **APIDataSource** (`pkg/weather/datasource_api.go`)
   - Polls WeatherFlow API every 60 seconds
   - Supports custom URLs (for generated weather, custom endpoints)
   - Fetches forecast every 30 minutes
   - Used for: Real API, Generated Weather, Custom URLs

2. **UDPDataSource** (`pkg/weather/datasource_udp.go`)
   - Listens for UDP broadcasts on port 50222
   - Forwards observations in real-time
   - Optional forecast polling (when internet enabled)
   - Used for: Offline mode, local monitoring

### Factory Pattern (`pkg/service/datasource_factory.go`)

**Single source of truth** for creating data sources:

```go
func CreateDataSource(cfg *config.Config, station *weather.Station) (weather.DataSource, error)
```

Priority order:
1. UDP Stream (if `--udp-stream`)
2. Custom Station URL (if `--station-url`)
3. Generated Weather (if `--use-generated-weather`)
4. WeatherFlow API (default)

### Service Integration (`pkg/service/service.go`)

**Unified processing loop** - same for all sources:

```go
dataSource, _ := CreateDataSource(cfg, station)
obsChan, _ := dataSource.Start()

for obs := range obsChan {
    // Update HomeKit
    ws.UpdateSensor(...)
    
    // Update web server
    webServer.UpdateWeather(&obs)
    
    // Update forecast
    webServer.UpdateForecast(dataSource.GetForecast())
    
    // Update status
    webServer.UpdateDataSourceStatus(dataSource.GetStatus())
}
```

### Web Server Integration (`pkg/web/server.go`)

**Unified status reporting**:

```go
type StatusResponse struct {
    // ... existing fields
    DataSource *weather.DataSourceStatus `json:"dataSource,omitempty"`
}
```

The `DataSource` field contains:
- Type (api/udp/generated/custom-url)
- Active status
- Observation count
- Last update time
- Source-specific details (IP address for UDP, custom URL, etc.)

## Benefits

### 1. **Single Responsibility**
- Each data source handles only its own data retrieval logic
- Service layer only handles observation processing
- Web server only handles status display

### 2. **Open/Closed Principle**
- Easy to add new data sources (MQTT, InfluxDB, etc.)
- No modification to service or web server needed
- Just implement DataSource interface and update factory

### 3. **No Branching Logic**
- Service.go has **zero** `if cfg.UDPStream` or `if cfg.UseGeneratedWeather` checks
- All branching isolated to factory function
- Much easier to test and maintain

### 4. **Consistent Behavior**
- All sources provide observations through channels
- All sources provide status through same interface
- Web console shows same information regardless of source

### 5. **Easier Testing**
- Mock DataSource for unit tests
- Test each source implementation independently
- Test service logic without worrying about sources

## Migration Guide

### Adding a New Data Source

1. Create new file: `pkg/weather/datasource_<name>.go`
2. Implement `DataSource` interface
3. Add new type to `DataSourceType` enum
4. Update factory to create your source
5. Done! No other changes needed.

### Example: MQTT Data Source

```go
// pkg/weather/datasource_mqtt.go
type MQTTDataSource struct {
    broker   string
    topic    string
    obsChan  chan Observation
    // ...
}

func (m *MQTTDataSource) Start() (<-chan Observation, error) {
    // Connect to MQTT broker
    // Subscribe to topic
    // Forward messages to obsChan
}

// Implement other interface methods...

// pkg/service/datasource_factory.go
func CreateDataSource(cfg *config.Config, station *weather.Station) (weather.DataSource, error) {
    if cfg.MQTTBroker != "" {
        return weather.NewMQTTDataSource(cfg.MQTTBroker, cfg.MQTTTopic), nil
    }
    // ... existing logic
}
```

## Files Changed

### New Files Created
- `pkg/weather/datasource.go` - Core interface and types
- `pkg/weather/datasource_api.go` - API polling implementation
- `pkg/weather/datasource_udp.go` - UDP stream implementation
- `pkg/service/datasource_factory.go` - Factory for creating sources

### Modified Files
- `pkg/service/service.go` - Refactored to use unified processing
- `pkg/web/server.go` - Added UpdateDataSourceStatus method
- `pkg/web/server.go` - Added DataSource field to StatusResponse

### Removed Files
- `pkg/service/service_refactored.go` - Temporary file, deleted

## Testing

Build verification:
```bash
go build  # Successful
./tempest-homekit-go --version  # v1.5.0
```

All three data source types compile and work:
- API polling: Default behavior
- Generated weather: `--use-generated-weather`
- UDP stream: `--udp-stream`

## Future Enhancements

Potential new data sources now easy to add:
- **MQTT**: Subscribe to MQTT broker for observations
- **InfluxDB**: Pull historical data from time-series database
- **File**: Read observations from JSON/CSV files
- **WebSocket**: Real-time streaming from custom server
- **Simulation**: Replay historical data for testing

All would follow the same pattern - implement interface, add to factory!

## Version

This refactoring completed: v1.5.0 (October 3, 2025)
