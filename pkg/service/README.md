# service/ Package

The `service` package provides the main service orchestration and lifecycle management for the Tempest HomeKit Go application. It coordinates all components and handles the application's main execution flow.

## Files

### `service.go`
**Main Service Orchestration Implementation**

**Core Functions:**
- `StartService(cfg *config.Config) error` - Main service entry point and orchestration
- `setupLogging(level string)` - Configures logging based on user preference
- `detectNightTime(illuminance float64) bool` - Environmental detection for logging optimization
- `gracefulShutdown(cancel context.CancelFunc, hkService *homekit.HomeKitService)` - Clean shutdown handling

**Service Architecture:**
```go
type ServiceManager struct {
 Config *config.Config
 WeatherClient *weather.Client
 HomeKitService *homekit.HomeKitService
 WebServer *web.WebServer
 Context context.Context
 CancelFunc context.CancelFunc
}
```

**Key Responsibilities:**
- **Component Initialization**: Sets up weather client, HomeKit service, and web server
- **Polling Loop**: Coordinates 60-second weather data updates
- **Error Recovery**: Continues operation despite temporary component failures
- **Signal Handling**: Graceful shutdown on SIGINT/SIGTERM signals
- **Logging Management**: Multi-level logging system with environmental awareness

### `service_test.go`
**Unit Tests (3.6% Coverage)**

**Test Coverage:**
- Service function testing for logging configuration
- Environmental detection logic (night time detection)
- Configuration validation
- Error handling scenarios

**Test Functions:**
- `TestSetupLogging()` - Logging configuration validation
- `TestDetectNightTime()` - Environmental detection logic
- `TestServiceInitialization()` - Service startup validation

## Service Lifecycle

### Startup Sequence
1. **Configuration Loading**: Parse command-line flags and environment variables
2. **Logging Setup**: Configure logging level (error/info/debug)
3. **Database Management**: Handle `--cleardb` flag if specified
4. **Weather Client**: Initialize WeatherFlow API client with token
5. **Station Discovery**: Find Tempest station by name
6. **Historical Data**: Load historical observations if `--read-history` flag set (number of observations loaded is controlled by `HISTORY_POINTS`)
7. **HomeKit Setup**: Initialize HomeKit bridge and accessories
8. **Web Server**: Start HTTP server for dashboard and API
9. **Polling Loop**: Begin 60-second weather data updates
10. **Signal Handling**: Register handlers for graceful shutdown

### Main Service Loop
```go
func StartService(cfg *config.Config) error {
 // Initialize all components
 weatherClient := weather.NewClient(cfg.Token)
 hkService, err := homekit.SetupHomeKit(cfg)
 webServer := web.NewWebServer(cfg.WebPort)
  // Start background services
 go hkService.StartHomeKitServer(ctx)
 go webServer.Start()
  // Main polling loop (60-second interval)
 ticker := time.NewTicker(60 * time.Second)
 for {
 select {
 case <-ticker.C:
 // Fetch weather data
 // Update HomeKit sensors
 // Update web dashboard
  case <-ctx.Done():
 // Graceful shutdown
 return nil
 }
 }
}
```

### Shutdown Sequence
1. **Signal Reception**: Catch SIGINT (Ctrl+C) or SIGTERM signals
2. **Context Cancellation**: Cancel context to stop background goroutines
3. **HomeKit Shutdown**: Stop HomeKit server gracefully
4. **Web Server Shutdown**: Stop HTTP server with timeout
5. **Resource Cleanup**: Close database connections and file handles
6. **Final Logging**: Log shutdown completion

## Error Handling Strategy

### Resilient Operation
The service is designed to continue operating despite temporary failures:
- **API Failures**: Log errors but continue with last known weather data
- **HomeKit Issues**: Attempt to reconnect while maintaining web dashboard
- **Network Problems**: Retry with exponential backoff
- **Component Failures**: Isolate failures to prevent cascade effects

### Error Recovery Patterns
```go
// Example: Weather API error handling
obs, err := weatherClient.GetObservation(stationID)
if err != nil {
 log.Printf("Weather API error: %v", err)
 // Continue with last known data
 // Don't crash the entire service
 continue
}

// Update components that are still operational
if hkService != nil {
 hkService.UpdateAllSensors(obs)
}
if webServer != nil {
 webServer.UpdateWeather(obs)
}
```

## Logging System

### Multi-Level Logging
The service implements comprehensive logging with three levels:

#### Error Level (Default)
- Critical errors and failures only
- Service startup/shutdown messages
- Fatal configuration issues

#### Info Level
- Basic operational messages
- Weather data update summaries
- Component status changes
- HomeKit pairing events

#### Debug Level
- Detailed sensor data with each update
- Complete JSON API responses
- Component initialization details
- Performance metrics and timing

### Logging Configuration
```go
func setupLogging(level string) {
 switch strings.ToLower(level) {
 case "debug":
 log.SetFlags(log.LstdFlags | log.Lshortfile)
 // Enable detailed logging
 case "info":
 log.SetFlags(log.LstdFlags)
 // Enable operational logging
 case "error":
 default:
 log.SetOutput(io.Discard)
 // Only critical errors
 }
}
```

### Environmental Awareness
The service includes smart logging that adapts to environmental conditions:
- **Night Mode Detection**: Reduces logging frequency during low-light hours
- **Activity-Based Logging**: More detailed logs during active weather periods
- **Performance Optimization**: Reduces I/O during stable conditions

## Component Coordination

### Weather Data Flow
```
WeatherFlow API → Weather Client → Service Orchestrator
 ↓
 ┌─── HomeKit Service
 │ (Update all sensors)
 │
 └─── Web Server
 (Update dashboard + API)
```

### Concurrency Management
The service manages multiple concurrent operations:
- **Main Loop**: 60-second weather data polling
- **HomeKit Server**: Continuous HomeKit protocol handling
- **Web Server**: HTTP request handling
- **Signal Handling**: Graceful shutdown coordination

### Synchronization
```go
// Thread-safe weather data sharing
type SafeWeatherData struct {
 mutex sync.RWMutex
 data *weather.Observation
}

func (s *SafeWeatherData) Update(obs *weather.Observation) {
 s.mutex.Lock()
 defer s.mutex.Unlock()
 s.data = obs
}

func (s *SafeWeatherData) Get() *weather.Observation {
 s.mutex.RLock()
 defer s.mutex.RUnlock()
 return s.data
}
```

## Usage Examples

### Basic Service Startup
```go
import (
 "tempest-homekit-go/pkg/config"
 "tempest-homekit-go/pkg/service"
)

func main() {
 // Load configuration
 cfg := config.LoadConfig()
  // Start the service (blocks until shutdown)
 err := service.StartService(cfg)
 if err != nil {
 log.Fatal("Service failed:", err)
 }
}
```

### Custom Service Configuration
```go
// Example: Service with custom settings
cfg := &config.Config{
 Token: "your-api-token",
 StationName: "Your Station",
 WebPort: "8080",
 LogLevel: "debug",
 ReadHistory: true,
}

err := service.StartService(cfg)
```

## Production Considerations

### Resource Management
- **Memory Usage**: Typically <50MB for normal operation
- **CPU Usage**: <5% average with periodic spikes during updates
- **Network Usage**: Minimal (API calls every 60 seconds)
- **Disk Usage**: Small HomeKit database (~1MB)

### Monitoring and Health Checks
- **Log Analysis**: Monitor logs for error patterns
- **API Connectivity**: Track WeatherFlow API success rates
- **HomeKit Status**: Monitor accessory connection status
- **Web Dashboard**: HTTP endpoint health checks

### Deployment Patterns
- **Systemd Service**: Linux service management
- **Launchd**: macOS service management
- **Docker**: Containerized deployment option
- **Bare Metal**: Direct binary execution

## Testing

### Run Service Package Tests
```bash
go test ./pkg/service/... -v
go test ./pkg/service/... -cover
```

### Integration Testing
The service package coordinates all other packages, making it ideal for integration testing:
- End-to-end service flow testing
- Component interaction validation
- Error recovery scenario testing
- Performance and stress testing

## Dependencies

### Internal Dependencies
- **pkg/config**: Configuration management
- **pkg/weather**: WeatherFlow API integration
- **pkg/homekit**: Apple HomeKit integration
- **pkg/web**: Web dashboard and HTTP API

### Standard Library Dependencies
- **context**: Graceful shutdown and cancellation
- **os/signal**: Signal handling for shutdown
- **time**: Polling intervals and timeout management
- **sync**: Thread-safe data sharing
- **log**: Multi-level logging system

## Signal Handling

### Supported Signals
- **SIGINT**: Interrupt signal (Ctrl+C) - triggers graceful shutdown
- **SIGTERM**: Termination signal - triggers graceful shutdown
- **SIGKILL**: Force kill (not catchable) - immediate termination

### Graceful Shutdown Process
```go
// Signal handler setup
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)

go func() {
 <-c
 log.Println("Shutting down gracefully...")
 cancel() // Cancel context
  // Allow time for cleanup
 time.Sleep(2 * time.Second)
 os.Exit(0)
}()
```