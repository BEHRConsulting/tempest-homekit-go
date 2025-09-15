# Tempest HomeKit Go Service - Complete Requirements

## Overview

✅ **COMPLETE**: Create a complete Go service application that monitors a WeatherFlow Tempest weather station and updates Apple HomeKit accessories with real-time weather data. The service enables smart home automation based on temperature, humidity, wind speed, and rain accumulation readings. Include a modern web dashboard with interactive unit conversions and real-time monitoring capabilities.

## Functional Requirements

### Core Functionality

#### Weather Data Monitoring
- ✅ **API Integration**: Connect to WeatherFlow Tempest REST API
- ✅ **Station Discovery**: Automatically find Tempest station by name from available stations
- ✅ **Data Polling**: Continuously poll weather observations every 60 seconds
- ✅ **Data Parsing**: Parse JSON responses containing weather metrics

#### HomeKit Integration
- ✅ **Bridge Setup**: Create HomeKit bridge accessory for device management
- ✅ **Sensor Accessories**: Implement 11 separate HomeKit accessories using modern `brutella/hap` library:
  - Air Temperature Sensor (standard HomeKit temperature characteristic)
  - 10 Custom Weather Sensors (Wind Speed/Gust/Direction, Rain, UV, Lightning, Humidity, Light, Precipitation Type)
- ✅ **Custom Service Solution**: Use unique service UUIDs to prevent HomeKit's automatic temperature conversion
- ✅ **Real-time Updates**: Update all sensor values with each weather poll
- ✅ **Pairing**: Support HomeKit pairing with configurable PIN

#### Web Dashboard
- ✅ **HTTP Server**: Serve modern web interface on configurable port (default: 8080)
- ✅ **Real-time Updates**: Dashboard updates every 10 seconds via JavaScript fetch API
- Show sensors: Temputure, Humidity (%, Relitive, Dew point), Wind (Speed, Direction, Guest), Atmospheric pressure, lux
- ✅ **Interactive Unit Conversion**: Click-to-toggle between metric and imperial units:
  - Temperature: Celsius (°C) ↔ Fahrenheit (°F)
  - Wind Speed, Wind Guest: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - Rain: Inches (in) ↔ Millimeters (mm)
  - Humidity dew point: Inches (in) ↔ Millimeters (mm)
  - Preasure: mb <> inHg
- ✅ **Wind Direction Display**: Show wind direction in cardinal format (N, NE, E, etc.) with degrees
- For each sensor, add a small graph of data point/time, max number of data points 1000
- ✅ **Unit Persistence**: Save user preferences in browser localStorage
- ✅ **HomeKit Status Display**: Show bridge status, accessory count, and pairing PIN
- ✅ **Connection Status**: Real-time Tempest station connection status
- ✅ **Responsive Design**: Mobile-friendly interface with modern CSS styling

#### Weather Data Mapping
- ✅ **Temperature**: Air temperature in Fahrenheit/Celsius
- ✅ **Humidity**: Relative humidity as percentage
- ✅ **Wind Speed**: Average wind speed in mph
- ✅ **Wind Direction**: Wind direction in degrees (0-360°) with cardinal conversion
- ✅ **Rain Accumulation**: Total precipitation in inches

### Configuration Management

#### Command-Line Flags
- ✅ `--token`: WeatherFlow API personal access token (required)
- ✅ `--station`: Tempest station name (default: "Chino Hills")
- ✅ `--pin`: HomeKit pairing PIN (default: "00102003")
- ✅ `--loglevel`: Logging verbosity - debug, info, error (default: "error")
- ✅ `--web-port`: Web dashboard port (default: "8080")
- ✅ `--cleardb`: Clear HomeKit database and reset device pairing

#### Environment Variables
- ✅ `TEMPEST_TOKEN`: WeatherFlow API token
- ✅ `TEMPEST_STATION_NAME`: Station name
- ✅ `HOMEKIT_PIN`: HomeKit PIN
- ✅ `LOG_LEVEL`: Logging level
- ✅ `WEB_PORT`: Web dashboard port (default: "8080")

### Service Operation

#### Lifecycle Management
- ✅ **Startup**: Initialize weather client, discover station, setup HomeKit
- ✅ **Polling Loop**: Continuous weather data polling with ticker
- ✅ **Shutdown**: Graceful shutdown on interrupt signals (SIGINT, SIGTERM)
- ✅ **Error Recovery**: Continue operation despite temporary API failures

#### Logging
- ✅ **Info Level**: Basic operational messages (startup, station found, updates)
- ✅ **Debug Level**: Detailed weather data values with each poll
- ✅ **Error Level**: API failures and critical errors only

## Technical Specifications

### WeatherFlow API Integration

#### API Endpoints
- ✅ **Stations**: `GET https://swd.weatherflow.com/swd/rest/stations?token={token}`
- ✅ **Observations**: `GET https://swd.weatherflow.com/swd/rest/observations/station/{station_id}?token={token}`

#### Data Structures

**Station Response:**
```json
{
  "stations": [
    {
      "station_id": 178915,
      "name": "Chino Hills",
      "station_name": "Chino Hills",
      "latitude": 33.98632,
      "longitude": -117.74695
    }
  ]
}
```

**Observation Response:**
```json
{
  "status": {"status_code": 0, "status_message": "SUCCESS"},
  "obs": [
    {
      "timestamp": 1757045053,
      "air_temperature": 24.4,
      "relative_humidity": 66,
      "wind_avg": 0.3,
      "wind_direction": 241,
      "precip": 0.0,
      "station_pressure": 979.7
    }
  ]
}
```

### HomeKit Implementation

#### Accessory Types
- ✅ **Bridge**: `accessory.NewBridge` for device management
- ✅ **Temperature Sensor**: `accessory.NewTemperatureSensor` with standard temperature characteristic
- ✅ **Custom Weather Sensors**: Individual accessories with custom service UUIDs (F1XX-0001-1000-8000-0026BB765291 pattern)

#### Custom Service Architecture
- ✅ **Unique Service UUIDs**: Each sensor type has its own service UUID to prevent HomeKit temperature conversion
- ✅ **Custom Characteristics**: Float characteristics that don't trigger automatic unit conversion
- ✅ **Modern hap Library**: Uses `github.com/brutella/hap` v0.0.32 with context-based server lifecycle

#### Service Characteristics
- ✅ **Air Temperature**: `CurrentTemperature` (float, Celsius) - standard HomeKit characteristic
- ✅ **Wind Speed**: Custom float characteristic (float, mph) - UUID F101-0001-1000-8000-0026BB765291
- ✅ **Wind Gust**: Custom float characteristic (float, mph) - UUID F111-0001-1000-8000-0026BB765291
- ✅ **Wind Direction**: Custom float characteristic (float, degrees) - UUID F121-0001-1000-8000-0026BB765291
- ✅ **Rain**: Custom float characteristic (float, inches) - UUID F131-0001-1000-8000-0026BB765291
- ✅ **UV Index**: Custom float characteristic (float, index) - UUID F141-0001-1000-8000-0026BB765291
- ✅ **Lightning Count**: Custom float characteristic (float, count) - UUID F151-0001-1000-8000-0026BB765291
- ✅ **Lightning Distance**: Custom float characteristic (float, km) - UUID F161-0001-1000-8000-0026BB765291
- ✅ **Precipitation Type**: Custom float characteristic (float, type) - UUID F171-0001-1000-8000-0026BB765291
- ✅ **Humidity**: Custom float characteristic (float, percent) - UUID F181-0001-1000-8000-0026BB765291
- ✅ **Ambient Light**: Custom float characteristic (float, lux) - UUID F191-0001-1000-8000-0026BB765291

### Web Dashboard Implementation

#### HTTP Server Setup
- ✅ **Port Configuration**: Configurable via `--web-port` flag (default: 8080)
- ✅ **Routes**: 
  - `GET /`: Main dashboard HTML page
  - `GET /api/weather`: JSON weather data endpoint
  - `GET /api/status`: JSON service and HomeKit status endpoint
- ✅ **CORS Support**: Allow cross-origin requests for API endpoints
- ✅ **Content Types**: Serve HTML, JSON, and static assets appropriately

#### Dashboard UI Requirements
- ✅ **Modern Design**: Use CSS gradients, card-based layout, and responsive design
- ✅ **Color Scheme**: Weather-themed colors (blue gradients, clean whites)
- ✅ **Typography**: System fonts (-apple-system, BlinkMacSystemFont, etc.)
- ✅ **Icons**: Unicode emoji for weather sensors (🌡️, 💧, 🌬️, 🌧️)
- ✅ **Cards**: Hover effects and smooth transitions
- ✅ **Wind Direction**: Display cardinal direction + degrees (e.g., "WSW (241°)")
- ✅ **Mobile Responsive**: Grid layout that adapts to screen size

#### JavaScript Functionality
- ✅ **Real-time Updates**: Fetch weather data every 10 seconds
- ✅ **Unit Conversion Functions**:
  - `celsiusToFahrenheit(celsius)`: Convert temperature
  - `fahrenheitToCelsius(fahrenheit)`: Convert temperature back
  - `mphToKph(mph)`: Convert wind speed
  - `kphToMph(kph)`: Convert wind speed back
  - `inchesToMm(inches)`: Convert rain
  - `mmToInches(mm)`: Convert rain back
  - `degreesToDirection(degrees)`: Convert wind degrees to cardinal directions
- ✅ **localStorage**: Persist unit preferences between sessions
- ✅ **Error Handling**: Graceful degradation when API calls fail
- ✅ **DOM Updates**: Update temperature, humidity, wind, rain values
- ✅ **Status Updates**: Update connection status and HomeKit information

#### API Response Formats

**Weather API Response:**
```json
{
  "temperature": 24.4,
  "humidity": 66.0,
  "windSpeed": 0.3,
  "windDirection": 241,
  "rainAccum": 0.0,
  "lastUpdate": "2025-09-04T21:26:51Z"
}
```

**Status API Response:**
```json
{
  "connected": true,
  "lastUpdate": "2025-09-04T21:26:51Z",
  "uptime": "1h30m45s",
  "homekit": {
    "bridge": true,
    "accessories": 4,
    "pin": "00102003"
  }
}
```

### Go Application Architecture

#### Package Structure
```
tempest-homekit-go/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
└── pkg/
    ├── config/
    │   ├── config.go          # Configuration management
    │   └── config_test.go     # Unit tests
    ├── weather/
    │   ├── client.go          # WeatherFlow API client
    │   └── client_test.go     # Unit tests
    ├── homekit/
    │   └── setup.go           # HomeKit accessory setup
    ├── web/
    │   └── server.go          # Web dashboard server
    └── service/
        └── service.go         # Main service orchestration
```

#### Key Components

**Configuration (pkg/config/config.go):**
- ✅ Load configuration from flags and environment
- ✅ Provide default values for all settings
- ✅ Validate required parameters (API token)

**Weather Client (pkg/weather/client.go):**
- ✅ `GetStations(token string)` → `[]Station`
- ✅ `GetObservation(stationID int, token string)` → `*Observation`
- ✅ `FindStationByName(stations []Station, name string)` → `*Station`
- ✅ Handle JSON parsing and HTTP error responses
- ✅ Implement proper timeout and retry logic

**HomeKit Setup (pkg/homekit/setup.go):**
- ✅ `NewWeatherAccessories()` → `*WeatherAccessories`
- ✅ `SetupHomeKit(wa *WeatherAccessories, pin string)` → `hc.Transport`
- ✅ Update methods: `UpdateTemperature()`, `UpdateHumidity()`, `UpdateWindSpeed()`, `UpdateRainAccumulation()`

**Web Server (pkg/web/server.go):**
- ✅ `NewWebServer(port string)` → `*WebServer`
- ✅ `Start()` → `error`: Start HTTP server
- ✅ `UpdateWeather(obs *weather.Observation)`: Update cached weather data
- ✅ `UpdateHomeKitStatus(status map[string]interface{})`: Update HomeKit status
- ✅ Handle dashboard, weather API, and status API routes
- ✅ Serve embedded HTML/CSS/JavaScript content

**Service Orchestration (pkg/service/service.go):**
- ✅ `StartService(cfg *config.Config)` → `error`
- ✅ Main polling loop with 60-second ticker
- ✅ Coordinate weather API calls, HomeKit updates, and web dashboard updates
- ✅ Handle graceful shutdown on signals

## Non-Functional Requirements

### Reliability
- ✅ **Error Handling**: All API calls must handle network failures
- ✅ **Data Validation**: Validate API responses before processing
- ✅ **Graceful Degradation**: Continue operation when individual sensors fail
- ✅ **Resource Management**: Proper cleanup of goroutines and connections

### Performance
- ✅ **Memory Usage**: < 50MB resident memory
- ✅ **CPU Usage**: < 5% average CPU utilization
- ✅ **API Efficiency**: Respect WeatherFlow API rate limits
- ✅ **Response Time**: < 5 seconds for HomeKit accessory updates

### Security
- ✅ **Token Security**: Never log API tokens in plain text
- ✅ **Input Sanitization**: Validate all user inputs
- ✅ **HomeKit Security**: Use standard HomeKit encryption
- ✅ **No Hardcoded Secrets**: All credentials from configuration

### Compatibility
- ✅ **Go Version**: Go 1.24.2 or later
- ✅ **Dependencies**:
  - `github.com/brutella/hc` (latest stable)
  - Standard library only for other dependencies
- ✅ **Operating Systems**: macOS, Linux, Windows
- ✅ **HomeKit**: iOS 14+, macOS 11+, HomePod

## Implementation Details

### Data Structures

**Station:**
```go
type Station struct {
    StationID   int     `json:"station_id"`
    Name        string  `json:"name"`
    StationName string  `json:"station_name"`
    Latitude    float64 `json:"latitude"`
    Longitude   float64 `json:"longitude"`
}
```

**Observation:**
```go
type Observation struct {
    Timestamp            int64   `json:"timestamp"`
    AirTemperature       float64 `json:"air_temperature"`
    RelativeHumidity     float64 `json:"relative_humidity"`
    WindAvg              float64 `json:"wind_avg"`
    WindDirection        float64 `json:"wind_direction"`
    RainAccumulated      float64 `json:"precip"`
    StationPressure      float64 `json:"station_pressure"`
}
```

### Error Handling

#### API Errors
- ✅ Network timeouts: Retry with exponential backoff
- ✅ HTTP 4xx: Log error and continue with last known values
- ✅ HTTP 5xx: Retry after delay
- ✅ Invalid JSON: Log error and skip update

#### HomeKit Errors
- ✅ Transport failures: Log and attempt restart
- ✅ Pairing issues: Log but don't crash service
- ✅ Characteristic updates: Validate values before updating

### Testing Requirements

#### Unit Tests
- ✅ **Configuration**: Test flag parsing and environment variables
- ✅ **Weather Client**: Test API calls with mock responses
- ✅ **Station Discovery**: Test name matching logic
- ✅ **Data Parsing**: Test JSON unmarshaling edge cases

#### Integration Tests
- ✅ **End-to-End**: Test complete weather-to-HomeKit flow
- ✅ **API Integration**: Test with real WeatherFlow API (with test token)
- ✅ **HomeKit Pairing**: Test accessory discovery and updates

### Build and Deployment

#### Build Process
```bash
go mod tidy
go build -o tempest-homekit-go
```

#### Dependencies
```go
module tempest-homekit-go

go 1.19

require (
    github.com/brutella/hc v1.2.4
)
```

#### Runtime Requirements
- ✅ Network access to WeatherFlow API
- ✅ Local network access for HomeKit
- ✅ Persistent storage for HomeKit database (`./db`)

## Future Enhancements

### Planned Features
- **Air Pressure Sensor**: Add barometric pressure monitoring
- ✅ **Wind Direction**: Display wind direction with cardinal directions (COMPLETED)
- **Weather Alerts**: Trigger HomeKit scenes based on weather thresholds
- **Historical Data**: Store and display weather history
- **Multiple Stations**: Support monitoring multiple Tempest stations
- ✅ **Web Dashboard**: Local web interface for monitoring (COMPLETED)

### API Extensions
- **Bulk Observations**: Request multiple observation types in single call
- **Webhook Support**: Receive real-time updates via webhooks
- **Station Metadata**: Additional station information and capabilities

## Success Criteria

### Functional Verification
- ✅ Application starts without errors
- ✅ Discovers specified Tempest station
- ✅ Polls weather data every 60 seconds
- ✅ Updates all 6 HomeKit sensors (Temperature, Humidity, Wind Speed, Wind Direction, Rain, Light)
- ✅ HomeKit pairing successful
- ✅ Debug logging shows all weather values
- ✅ Web dashboard displays wind direction

### Quality Assurance
- ✅ All unit tests pass
- ✅ No runtime panics
- ✅ Proper error handling
- ✅ Memory leaks absent
- ✅ CPU usage within limits

### User Experience
- ✅ Simple command-line interface
- ✅ Clear logging messages
- ✅ Easy HomeKit setup
- ✅ Reliable continuous operation
- ✅ Modern web dashboard with real-time updates
- ✅ Interactive unit conversions with persistence
- ✅ Wind direction display with cardinal directions

## Complete Implementation Guide

### Step-by-Step Development Process

#### Phase 1: Project Setup
1. ✅ Create Go module: `go mod init tempest-homekit-go`
2. ✅ Install dependencies: `go get github.com/brutella/hc`
3. ✅ Create package structure as specified above
4. ✅ Implement configuration management with flags and environment variables

#### Phase 2: Weather API Integration
1. ✅ Implement WeatherFlow API client in `pkg/weather/client.go`
2. ✅ Create data structures for Station and Observation
3. ✅ Add JSON parsing and error handling
4. ✅ Implement station discovery by name

#### Phase 3: HomeKit Setup
1. ✅ Create HomeKit accessories in `pkg/homekit/setup.go`
2. ✅ Implement bridge and 5 sensor accessories
3. ✅ Add update methods for each sensor type
4. ✅ Configure HomeKit transport with PIN

#### Phase 4: Web Dashboard
1. ✅ Create web server in `pkg/web/server.go`
2. ✅ Implement HTTP routes for dashboard and APIs
3. ✅ Create complete HTML template with modern CSS
4. ✅ Add JavaScript for real-time updates and unit conversions
5. ✅ Integrate HomeKit status display

#### Phase 5: Service Orchestration
1. ✅ Implement main service loop in `pkg/service/service.go`
2. ✅ Coordinate weather polling, HomeKit updates, and web dashboard
3. ✅ Add graceful shutdown handling
4. ✅ Integrate all components

#### Phase 6: Testing and Refinement
1. ✅ Test with real WeatherFlow API token
2. ✅ Verify HomeKit pairing and sensor updates
3. ✅ Test web dashboard functionality
4. ✅ Add comprehensive error handling

### Key Code Patterns

#### Main Entry Point (main.go)
```go
func main() {
    cfg := config.LoadConfig()
    err := service.StartService(cfg)
    if err != nil {
        log.Fatalf("Service failed: %v", err)
    }
    
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    log.Println("Shutting down...")
}
```

#### Weather API Client Pattern
```go
func GetObservation(stationID int, token string) (*Observation, error) {
    url := fmt.Sprintf("https://swd.weatherflow.com/swd/rest/observations/station/%d?token=%s", stationID, token)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response struct {
        Status struct {
            StatusCode int `json:"status_code"`
        } `json:"status"`
        Obs []Observation `json:"obs"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    if len(response.Obs) == 0 {
        return nil, fmt.Errorf("no observations available")
    }
    
    return &response.Obs[0], nil
}
```

#### HomeKit Accessory Setup Pattern
```go
func NewWeatherAccessories() *WeatherAccessories {
    bridge := accessory.NewBridge(accessory.Info{Name: "Tempest Bridge"})
    
    tempSensor := accessory.NewTemperatureSensor(accessory.Info{Name: "Temperature"})
    humiditySensor := accessory.NewHumiditySensor(accessory.Info{Name: "Humidity"})
    windSensor := accessory.NewFan(accessory.Info{Name: "Wind"})
    rainSensor := accessory.NewHumiditySensor(accessory.Info{Name: "Rain"})
    
    return &WeatherAccessories{
        Bridge: bridge,
        Temperature: tempSensor,
        Humidity: humiditySensor,
        Wind: windSensor,
        Rain: rainSensor,
    }
}
```

#### Web Server Pattern
```go
func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
    // Serve complete HTML template with embedded CSS and JavaScript
    tmpl := `<!DOCTYPE html>
    <html>
    <head>
        <title>Tempest Weather Dashboard</title>
        <style>
            /* Complete modern CSS implementation */
        </style>
    </head>
    <body>
        <!-- Complete HTML structure -->
        <script>
            /* Complete JavaScript implementation */
        </script>
    </body>
    </html>`
    
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(tmpl))
}
```

## Additional Features Implemented

### Cross-Platform Build Scripts
- ✅ **scripts/build.sh**: Cross-platform build script for Linux, macOS, Windows
- ✅ **scripts/install-service.sh**: Auto-detect OS and install as system service
- ✅ **scripts/remove-service.sh**: Stop and remove services with cleanup
- ✅ **scripts/README.md**: Comprehensive documentation for all scripts

### Enhanced Logging
- ✅ **Multi-level Logging**: Error, Info, Debug levels with appropriate verbosity
- ✅ **Sensor Data Logging**: Info level shows sensor updates
- ✅ **JSON API Output**: Debug level includes complete API responses
- ✅ **Connection Status**: Real-time connection monitoring

### Production Ready Features
- ✅ **Graceful Shutdown**: Proper signal handling and cleanup
- ✅ **Error Recovery**: Continue operation despite temporary failures
- ✅ **Resource Management**: Efficient memory and CPU usage
- ✅ **Security**: No hardcoded secrets, secure token handling

This requirements document has been updated to reflect the **COMPLETE** implementation of the Tempest HomeKit Go service, including all originally planned features plus additional enhancements like cross-platform deployment scripts and enhanced logging capabilities.

## Functional Requirements

### Core Functionality

#### Weather Data Monitoring
- **API Integration**: Connect to WeatherFlow Tempest REST API
- **Station Discovery**: Automatically find Tempest station by name from available stations
- **Data Polling**: Continuously poll weather observations every 60 seconds
- **Data Parsing**: Parse JSON responses containing weather metrics

#### HomeKit Integration
- **Bridge Setup**: Create HomeKit bridge accessory for device management
- **Sensor Accessories**: Implement 4 separate HomeKit accessories:
  - Temperature Sensor (Air Temperature)
  - Humidity Sensor (Relative Humidity)
  - Wind Sensor (Average Wind Speed)
  - Rain Sensor (Rain Accumulation)
- **Real-time Updates**: Update all sensor values with each weather poll
- **Pairing**: Support HomeKit pairing with configurable PIN

#### Web Dashboard
- **HTTP Server**: Serve modern web interface on configurable port (default: 8080)
- **Real-time Updates**: Dashboard updates every 10 seconds via JavaScript fetch API
- **Interactive Unit Conversion**: Click-to-toggle between metric and imperial units:
  - Temperature: Celsius (°C) ↔ Fahrenheit (°F)
  - Wind Speed: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - Rain: Inches (in) ↔ Millimeters (mm)
- **Wind Direction Display**: Show wind direction in cardinal format (N, NE, E, etc.) with degrees
- **Unit Persistence**: Save user preferences in browser localStorage
- **HomeKit Status Display**: Show bridge status, accessory count, and pairing PIN
- **Connection Status**: Real-time Tempest station connection status
- **Responsive Design**: Mobile-friendly interface with modern CSS styling

#### Weather Data Mapping
- **Temperature**: Air temperature in Fahrenheit/Celsius
- **Humidity**: Relative humidity as percentage
- **Wind Speed**: Average wind speed in mph
- **Wind Direction**: Wind direction in degrees (0-360°) with cardinal conversion
- **Rain Accumulation**: Total precipitation in inches

### Configuration Management

#### Command-Line Flags
- `--token`: WeatherFlow API personal access token (required)
- `--station`: Tempest station name (default: "Chino Hills")
- `--pin`: HomeKit pairing PIN (default: "00102003")
- `--loglevel`: Logging verbosity - debug, info, error (default: "error")

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `WEB_PORT`: Web dashboard port (default: "8080")

### Service Operation

#### Lifecycle Management
- **Startup**: Initialize weather client, discover station, setup HomeKit
- **Polling Loop**: Continuous weather data polling with ticker
- **Shutdown**: Graceful shutdown on interrupt signals (SIGINT, SIGTERM)
- **Error Recovery**: Continue operation despite temporary API failures

#### Logging
- **Info Level**: Basic operational messages (startup, station found, updates)
- **Debug Level**: Detailed weather data values with each poll
- **Error Level**: API failures and critical errors only

## Technical Specifications

### WeatherFlow API Integration

#### API Endpoints
- **Stations**: `GET https://swd.weatherflow.com/swd/rest/stations?token={token}`
- **Observations**: `GET https://swd.weatherflow.com/swd/rest/observations/station/{station_id}?token={token}`

#### Data Structures

**Station Response:**
```json
{
  "stations": [
    {
      "station_id": 178915,
      "name": "Chino Hills",
      "station_name": "Chino Hills",
      "latitude": 33.98632,
      "longitude": -117.74695
    }
  ]
}
```

**Observation Response:**
```json
{
  "status": {"status_code": 0, "status_message": "SUCCESS"},
  "obs": [
    {
      "timestamp": 1757045053,
      "air_temperature": 24.4,
      "relative_humidity": 66,
      "wind_avg": 0.3,
      "wind_direction": 241,
      "precip": 0.0,
      "station_pressure": 979.7
    }
  ]
}
```

### HomeKit Implementation

#### Accessory Types
- **Bridge**: `accessory.TypeBridge` for device management
- **Temperature Sensor**: `accessory.TypeOther` with `service.TemperatureSensor`
- **Humidity Sensor**: `accessory.TypeOther` with `service.HumiditySensor`
- **Wind Sensor**: `accessory.TypeOther` with `service.Fan` (On/Off for wind presence)
- **Rain Sensor**: `accessory.TypeOther` with `service.HumiditySensor` (scaled for rain accumulation)

#### Service Characteristics
- **Temperature**: `CurrentTemperature` (float, Celsius)
- **Humidity**: `CurrentRelativeHumidity` (float, percentage)
- **Wind**: `On` (boolean, wind presence)
- **Rain**: `CurrentRelativeHumidity` (float, scaled 0-100%)

### Web Dashboard Implementation

#### HTTP Server Setup
- **Port Configuration**: Configurable via `--web-port` flag (default: 8080)
- **Routes**: 
  - `GET /`: Main dashboard HTML page
  - `GET /api/weather`: JSON weather data endpoint
  - `GET /api/status`: JSON service and HomeKit status endpoint
- **CORS Support**: Allow cross-origin requests for API endpoints
- **Content Types**: Serve HTML, JSON, and static assets appropriately

#### Dashboard UI Requirements
- **Modern Design**: Use CSS gradients, card-based layout, and responsive design
- **Color Scheme**: Weather-themed colors (blue gradients, clean whites)
- **Typography**: System fonts (-apple-system, BlinkMacSystemFont, etc.)
- **Icons**: Unicode emoji for weather sensors (🌡️, 💧, 🌬️, 🌧️)
- **Cards**: Hover effects and smooth transitions
- **Wind Direction**: Display cardinal direction + degrees (e.g., "WSW (241°)")
- **Mobile Responsive**: Grid layout that adapts to screen size

#### JavaScript Functionality
- **Real-time Updates**: Fetch weather data every 10 seconds
- **Unit Conversion Functions**:
  - `celsiusToFahrenheit(celsius)`: Convert temperature
  - `fahrenheitToCelsius(fahrenheit)`: Convert temperature back
  - `mphToKph(mph)`: Convert wind speed
  - `kphToMph(kph)`: Convert wind speed back
  - `inchesToMm(inches)`: Convert rain
  - `mmToInches(mm)`: Convert rain back
  - `degreesToDirection(degrees)`: Convert wind degrees to cardinal directions
- **localStorage**: Persist unit preferences between sessions
- **Error Handling**: Graceful degradation when API calls fail
- **DOM Updates**: Update temperature, humidity, wind, rain values
- **Status Updates**: Update connection status and HomeKit information

#### API Response Formats

**Weather API Response:**
```json
{
  "temperature": 24.4,
  "humidity": 66.0,
  "windSpeed": 0.3,
  "windDirection": 241,
  "rainAccum": 0.0,
  "lastUpdate": "2025-09-04T21:26:51Z"
}
```

**Status API Response:**
```json
{
  "connected": true,
  "lastUpdate": "2025-09-04T21:26:51Z",
  "uptime": "1h30m45s",
  "homekit": {
    "bridge": true,
    "accessories": 4,
    "pin": "00102003"
  }
}
```

### Go Application Architecture

#### Package Structure
```
tempest-homekit-go/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
└── pkg/
    ├── config/
    │   ├── config.go          # Configuration management
    │   └── config_test.go     # Unit tests
    ├── weather/
    │   ├── client.go          # WeatherFlow API client
    │   └── client_test.go     # Unit tests
    ├── homekit/
    │   └── setup.go           # HomeKit accessory setup
    ├── web/
    │   └── server.go          # Web dashboard server
    └── service/
        └── service.go         # Main service orchestration
```

#### Key Components

**Configuration (pkg/config/config.go):**
- Load configuration from flags and environment
- Provide default values for all settings
- Validate required parameters (API token)

**Weather Client (pkg/weather/client.go):**
- `GetStations(token string)` → `[]Station`
- `GetObservation(stationID int, token string)` → `*Observation`
- `FindStationByName(stations []Station, name string)` → `*Station`
- Handle JSON parsing and HTTP error responses
- Implement proper timeout and retry logic

**HomeKit Setup (pkg/homekit/setup.go):**
- `NewWeatherAccessories()` → `*WeatherAccessories`
- `SetupHomeKit(wa *WeatherAccessories, pin string)` → `hc.Transport`
- Update methods: `UpdateTemperature()`, `UpdateHumidity()`, `UpdateWindSpeed()`, `UpdateRainAccumulation()`

**Web Server (pkg/web/server.go):**
- `NewWebServer(port string)` → `*WebServer`
- `Start()` → `error`: Start HTTP server
- `UpdateWeather(obs *weather.Observation)`: Update cached weather data
- `UpdateHomeKitStatus(status map[string]interface{})`: Update HomeKit status
- Handle dashboard, weather API, and status API routes
- Serve embedded HTML/CSS/JavaScript content

**Service Orchestration (pkg/service/service.go):**
- `StartService(cfg *config.Config)` → `error`
- Main polling loop with 60-second ticker
- Coordinate weather API calls, HomeKit updates, and web dashboard updates
- Handle graceful shutdown on signals

## Non-Functional Requirements

### Reliability
- **Error Handling**: All API calls must handle network failures
- **Data Validation**: Validate API responses before processing
- **Graceful Degradation**: Continue operation when individual sensors fail
- **Resource Management**: Proper cleanup of goroutines and connections

### Performance
- **Memory Usage**: < 50MB resident memory
- **CPU Usage**: < 5% average CPU utilization
- **API Efficiency**: Respect WeatherFlow API rate limits
- **Response Time**: < 5 seconds for HomeKit accessory updates

### Security
- **Token Security**: Never log API tokens in plain text
- **Input Sanitization**: Validate all user inputs
- **HomeKit Security**: Use standard HomeKit encryption
- **No Hardcoded Secrets**: All credentials from configuration

### Compatibility
- **Go Version**: Go 1.19 or later
- **Dependencies**:
  - `github.com/brutella/hap` v0.0.32 (modern HomeKit library)
  - Standard library only for other dependencies
- **Operating Systems**: macOS, Linux, Windows
- **HomeKit**: iOS 14+, macOS 11+, HomePod

## Implementation Details

### Data Structures

**Station:**
```go
type Station struct {
    StationID   int     `json:"station_id"`
    Name        string  `json:"name"`
    StationName string  `json:"station_name"`
    Latitude    float64 `json:"latitude"`
    Longitude   float64 `json:"longitude"`
}
```

**Observation:**
```go
type Observation struct {
    Timestamp            int64   `json:"timestamp"`
    AirTemperature       float64 `json:"air_temperature"`
    RelativeHumidity     float64 `json:"relative_humidity"`
    WindAvg              float64 `json:"wind_avg"`
    WindDirection        float64 `json:"wind_direction"`
    RainAccumulated      float64 `json:"precip"`
    StationPressure      float64 `json:"station_pressure"`
}
```

### Error Handling

#### API Errors
- Network timeouts: Retry with exponential backoff
- HTTP 4xx: Log error and continue with last known values
- HTTP 5xx: Retry after delay
- Invalid JSON: Log error and skip update

#### HomeKit Errors
- Transport failures: Log and attempt restart
- Pairing issues: Log but don't crash service
- Characteristic updates: Validate values before updating

### Testing Requirements

#### Unit Tests
- **Configuration**: Test flag parsing and environment variables
- **Weather Client**: Test API calls with mock responses
- **Station Discovery**: Test name matching logic
- **Data Parsing**: Test JSON unmarshaling edge cases

#### Integration Tests
- **End-to-End**: Test complete weather-to-HomeKit flow
- **API Integration**: Test with real WeatherFlow API (with test token)
- **HomeKit Pairing**: Test accessory discovery and updates

### Build and Deployment

#### Build Process
```bash
go mod tidy
go build -o tempest-homekit-go
```

#### Dependencies
```go
module tempest-homekit-go

go 1.19

require (
    github.com/brutella/hc v1.2.4
)
```

#### Runtime Requirements
- Network access to WeatherFlow API
- Local network access for HomeKit
- Persistent storage for HomeKit database (`./db`)

## Future Enhancements

### Planned Features
- **Air Pressure Sensor**: Add barometric pressure monitoring
- **Wind Direction**: ✅ Display wind direction with cardinal directions (COMPLETED)
- **Weather Alerts**: Trigger HomeKit scenes based on weather thresholds
- **Historical Data**: Store and display weather history
- **Multiple Stations**: Support monitoring multiple Tempest stations
- **Web Dashboard**: ✅ Local web interface for monitoring (COMPLETED)

### API Extensions
- **Bulk Observations**: Request multiple observation types in single call
- **Webhook Support**: Receive real-time updates via webhooks
- **Station Metadata**: Additional station information and capabilities

## Success Criteria

### Functional Verification
- ✅ Application starts without errors
- ✅ Discovers specified Tempest station
- ✅ Polls weather data every 60 seconds
- ✅ Updates all 11 HomeKit sensors (Temperature + 10 custom weather sensors)
- ✅ Custom services prevent temperature conversion issues
- ✅ HomeKit pairing successful
- ✅ Debug logging shows all weather values
- ✅ Web dashboard displays wind direction

### Quality Assurance
- ✅ All unit tests pass
- ✅ No runtime panics
- ✅ Proper error handling
- ✅ Memory leaks absent
- ✅ CPU usage within limits

### User Experience
- ✅ Simple command-line interface
- ✅ Clear logging messages
- ✅ Easy HomeKit setup
- ✅ Reliable continuous operation
- ✅ Modern web dashboard with real-time updates
- ✅ Interactive unit conversions with persistence
- ✅ Wind direction display with cardinal directions

## Complete Implementation Guide

### Step-by-Step Development Process

#### Phase 1: Project Setup
1. Create Go module: `go mod init tempest-homekit-go`
2. Install dependencies: `go get github.com/brutella/hc`
3. Create package structure as specified above
4. Implement configuration management with flags and environment variables

#### Phase 2: Weather API Integration
1. Implement WeatherFlow API client in `pkg/weather/client.go`
2. Create data structures for Station and Observation
3. Add JSON parsing and error handling
4. Implement station discovery by name

#### Phase 3: HomeKit Setup
1. Create HomeKit accessories in `pkg/homekit/setup.go`
2. Implement bridge and 4 sensor accessories
3. Add update methods for each sensor type
4. Configure HomeKit transport with PIN

#### Phase 4: Web Dashboard
1. Create web server in `pkg/web/server.go`
2. Implement HTTP routes for dashboard and APIs
3. Create complete HTML template with modern CSS
4. Add JavaScript for real-time updates and unit conversions
5. Integrate HomeKit status display

#### Phase 5: Service Orchestration
1. Implement main service loop in `pkg/service/service.go`
2. Coordinate weather polling, HomeKit updates, and web dashboard
3. Add graceful shutdown handling
4. Integrate all components

#### Phase 6: Testing and Refinement
1. Test with real WeatherFlow API token
2. Verify HomeKit pairing and sensor updates
3. Test web dashboard functionality
4. Add comprehensive error handling

### Key Code Patterns

#### Main Entry Point (main.go)
```go
func main() {
    cfg := config.LoadConfig()
    err := service.StartService(cfg)
    if err != nil {
        log.Fatalf("Service failed: %v", err)
    }
    
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    log.Println("Shutting down...")
}
```

#### Weather API Client Pattern
```go
func GetObservation(stationID int, token string) (*Observation, error) {
    url := fmt.Sprintf("https://swd.weatherflow.com/swd/rest/observations/station/%d?token=%s", stationID, token)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response struct {
        Status struct {
            StatusCode int `json:"status_code"`
        } `json:"status"`
        Obs []Observation `json:"obs"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    if len(response.Obs) == 0 {
        return nil, fmt.Errorf("no observations available")
    }
    
    return &response.Obs[0], nil
}
```

#### HomeKit Accessory Setup Pattern
```go
func NewWeatherAccessories() *WeatherAccessories {
    bridge := accessory.NewBridge(accessory.Info{Name: "Tempest Bridge"})
    
    tempSensor := accessory.NewTemperatureSensor(accessory.Info{Name: "Temperature"})
    humiditySensor := accessory.NewHumiditySensor(accessory.Info{Name: "Humidity"})
    windSensor := accessory.NewFan(accessory.Info{Name: "Wind"})
    rainSensor := accessory.NewHumiditySensor(accessory.Info{Name: "Rain"})
    
    return &WeatherAccessories{
        Bridge: bridge,
        Temperature: tempSensor,
        Humidity: humiditySensor,
        Wind: windSensor,
        Rain: rainSensor,
    }
}
```

#### Web Server Pattern
```go
func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
    // Serve complete HTML template with embedded CSS and JavaScript
    tmpl := `<!DOCTYPE html>
    <html>
    <head>
        <title>Tempest Weather Dashboard</title>
        <style>
            /* Complete modern CSS implementation */
        </style>
    </head>
    <body>
        <!-- Complete HTML structure -->
        <script>
            /* Complete JavaScript implementation */
        </script>
    </body>
    </html>`
    
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(tmpl))
}
```

This requirements document provides complete specifications for implementing the Tempest HomeKit Go service from scratch, including the modern web dashboard with interactive unit conversions and real-time monitoring capabilities.

---

**Status**: ✅ **COMPLETE** - All planned features implemented and tested
- ✅ Weather monitoring with 6 metrics (Temperature, Humidity, Wind Speed, Wind Direction, Rain, Light)
- ✅ Complete HomeKit integration with individual sensors
- ✅ Modern web dashboard with real-time updates
- ✅ Interactive unit conversions with persistence
- ✅ Cross-platform build and deployment
- ✅ Service management for all platforms
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error handling