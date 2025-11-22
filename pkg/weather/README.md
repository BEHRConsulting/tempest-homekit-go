# weather/ Package

The `weather` package provides a comprehensive client for the WeatherFlow Tempest API, handling station discovery, weather data retrieval, and data parsing for the Tempest HomeKit Go application.

## Files

### `client.go`
**WeatherFlow API Client Implementation**

**Core Functions:**
- `NewClient(token string) *Client` - Creates a new WeatherFlow API client
- `GetStations() ([]Station, error)` - Retrieves all available weather stations
- `GetObservation(stationID int) (*Observation, error)` - Gets latest weather observation
- `GetHistoricalObservations(stationID int, count int) ([]Observation, error)` - Loads historical data
- `FindStationByName(stations []Station, name string) *Station` - Finds station by name

**Data Structures:**
```go
type Station struct {
 StationID int `json:"station_id"`
 Name string `json:"name"`
 StationName string `json:"station_name"`
 Latitude float64 `json:"latitude"`
 Longitude float64 `json:"longitude"`
}

type Observation struct {
 Timestamp int64 `json:"timestamp"`
 AirTemperature float64 `json:"air_temperature"`
 RelativeHumidity float64 `json:"relative_humidity"`
 WindAvg float64 `json:"wind_avg"`
 WindGust float64 `json:"wind_gust"`
 WindDirection float64 `json:"wind_direction"`
 RainAccumulated float64 `json:"precip"`
 StationPressure float64 `json:"station_pressure"`
 UV int `json:"uv"`
 Illuminance float64 `json:"illuminance"`
 LightningCount int `json:"lightning_strike_count"`
 LightningDistance float64 `json:"lightning_strike_avg_distance"`
 PrecipitationType int `json:"precip_type"`
}
```

**Key Features:**
- **Robust HTTP Client**: Configurable timeout and retry logic
- **JSON Parsing**: Safe type conversion with error handling
- **Station Discovery**: Fuzzy matching for station name lookup
- **Historical Data**: Support for loading historical observations
- **Error Handling**: Comprehensive API error management
- **Rate Limiting**: Respects WeatherFlow API rate limits

### `client_test.go`
**Unit Tests (16.2% Coverage)**

**Test Coverage:**
- API client creation and configuration
- Station discovery and name matching
- JSON parsing utilities and helper functions
- Error handling for API failures
- Mock HTTP responses for testing

**Test Functions:**
- `TestNewClient()` - Client initialization
- `TestFindStationByName()` - Station name matching logic
- `TestParseObservation()` - JSON parsing validation
- `TestAPIErrorHandling()` - Error response handling
- `TestHistoricalDataParsing()` - Historical data processing

## WeatherFlow API Integration

### API Endpoints

#### Stations Endpoint
```
GET https://swd.weatherflow.com/swd/rest/stations?token={token}
```
**Response Structure:**
```json
{
 "stations": [
 {
 "station_id": 178915,
 "name": "Chino Hills",
 "latitude": 33.98632,
 "longitude": -117.74695
 }
 ]
}
```

#### Observations Endpoint
```
GET https://swd.weatherflow.com/swd/rest/observations/station/{station_id}?token={token}
```
**Response Structure:**
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
 "station_pressure": 979.7,
 "uv": 2,
 "illuminance": 15000.0
 }
 ]
}
```

### Historical Data Support
```
GET https://swd.weatherflow.com/swd/rest/observations/station/{station_id}?token={token}&time_start={timestamp}&time_end={timestamp}
```

## Usage Examples

### Basic Client Setup
```go
import "tempest-homekit-go/pkg/weather"

// Create client with API token
client := weather.NewClient("your-api-token")

// Discover stations
stations, err := client.GetStations()
if err != nil {
 log.Fatal("Failed to get stations:", err)
}

// Find specific station
station := weather.FindStationByName(stations, "Chino Hills")
if station == nil {
 log.Fatal("Station not found")
}
```

### Get Weather Data
```go
// Get latest observation
obs, err := client.GetObservation(station.StationID)
if err != nil {
 log.Fatal("Failed to get observation:", err)
}

fmt.Printf("Temperature: %.1f°C\n", obs.AirTemperature)
fmt.Printf("Humidity: %.1f%%\n", obs.RelativeHumidity)
fmt.Printf("Wind: %.1f mph from %d°\n", obs.WindAvg, int(obs.WindDirection))
```

### Historical Data Loading
```go
// Load 200 historical observations (for --history-read flag). The number of points loaded is controlled by HISTORY_POINTS.
historical, err := client.GetHistoricalObservations(station.StationID, 200)
if err != nil {
 log.Fatal("Failed to load historical data:", err)
}

fmt.Printf("Loaded %d historical observations\n", len(historical))
```

## Weather Metrics

### Supported Measurements
- **Air Temperature** - Ambient temperature (°C)
- **Relative Humidity** - Humidity percentage (0-100%)
- **Wind Speed** - Average wind speed (m/s)
- **Wind Gust** - Peak wind gust (m/s)
- **Wind Direction** - Wind direction (0-360°)
- **Rain Accumulation** - Precipitation (mm)
- **Station Pressure** - Barometric pressure (mb)
- **UV Index** - UV exposure level (0-11+)
- **Illuminance** - Ambient light level (lux)
- **Lightning Count** - Lightning strikes detected
- **Lightning Distance** - Average strike distance (km)
- **Precipitation Type** - Type of precipitation (integer code)

### Unit Conversions
The client receives data in metric units from the WeatherFlow API:
- **Temperature**: Celsius → Fahrenheit conversion in web dashboard
- **Wind Speed**: m/s → mph/kph conversion in web dashboard
- **Rain**: mm → inches conversion in web dashboard
- **Pressure**: mb (native) with inHg conversion option

## Error Handling

### API Error Types
- **Network Errors**: Connection timeouts, DNS failures
- **HTTP Errors**: 4xx client errors, 5xx server errors
- **JSON Parsing Errors**: Malformed response data
- **Data Validation Errors**: Missing or invalid fields

### Error Response Handling
```go
obs, err := client.GetObservation(stationID)
if err != nil {
 switch {
 case strings.Contains(err.Error(), "timeout"):
 // Handle timeout - retry with exponential backoff
 case strings.Contains(err.Error(), "404"):
 // Station not found - check station ID
 case strings.Contains(err.Error(), "401"):
 // Invalid API token - check credentials
 default:
 // Other errors - log and continue
 }
}
```

## Testing

### Run Weather Package Tests
```bash
go test ./pkg/weather/... -v
go test ./pkg/weather/... -cover
```

### Mock Testing
The tests use mock HTTP responses to simulate various API scenarios:
- Successful API responses
- Network failures
- Invalid JSON responses
- Empty observation data

## Rate Limiting

### WeatherFlow API Limits
- **Standard Rate**: 1,000 requests per day
- **Burst Rate**: 20 requests per minute
- **Historical Data**: Limited to avoid rate limit issues

### Best Practices
- **Polling Interval**: 60-second intervals for live data
- **Historical Loading**: Use 5-minute intervals to respect rate limits
- **Error Handling**: Exponential backoff on rate limit errors
- **Caching**: Cache observations to reduce API calls

## Dependencies

- **net/http**: HTTP client for API communication
- **encoding/json**: JSON parsing and unmarshaling
- **time**: Timestamp handling and timeout management
- **fmt**, **strings**: String formatting and manipulation
- **Standard Library**: Robust error handling and type conversion