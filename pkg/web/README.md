# web/ Package

The `web` package provides a comprehensive HTTP server and modern web dashboard for the Tempest HomeKit Go application. It serves real-time weather data, interactive charts, and provides REST API endpoints for weather monitoring.

## Files

### `server.go`
**HTTP Server and Web Dashboard Implementation**

**Core Functions:**
- `NewWebServer(port string) *WebServer` - Creates a new web server instance
- `Start() error` - Starts the HTTP server with all routes configured
- `UpdateWeather(obs *weather.Observation)` - Updates cached weather data
- `ServeStaticFile(filename string)` - Serves static assets with cache-busting
- `AnalyzePressureTrend(history []float64) (string, string, string)` - Pressure forecasting

**HTTP Routes:**
- `GET /` - Main dashboard HTML page
- `GET /api/weather` - JSON weather data endpoint
- `GET /api/status` - Service and HomeKit status endpoint
- `GET /pkg/web/static/` - Static assets (JavaScript, CSS, images)

**Dashboard Features:**
- **Real-time Updates**: JavaScript updates every 10 seconds
- **Interactive Charts**: Historical data visualization using Chart.js
- **Unit Conversions**: Click-to-toggle between metric/imperial units
- **Pressure Analysis**: Weather forecasting based on pressure trends
- **Tooltip System**: Information tooltips for all weather sensors
- **Responsive Design**: Mobile-friendly interface

### `server_test.go`
**Unit Tests (50.5% Coverage)**

**Test Coverage:**
- HTTP endpoint testing using `httptest.ResponseRecorder`
- Pressure analysis function validation
- Static file serving with cache-busting
- JSON API response validation
- Error handling for missing data

**Test Functions:**
- `TestWebServerCreation()` - Server initialization
- `TestWeatherEndpoint()` - Weather API endpoint testing
- `TestStatusEndpoint()` - Status API endpoint testing
- `TestPressureAnalysis()` - Pressure forecasting logic
- `TestStaticFileServing()` - Static asset serving

### `static/` Directory
**Frontend Assets for Web Dashboard**

## static/ Directory Contents

### `script.js`
**Main JavaScript Application (~800+ lines)**

**Core Features:**
- **Real-time Data Fetching**: Updates weather data every 10 seconds
- **Interactive Charts**: Chart.js integration for historical data visualization
- **Unit Conversion System**: Toggle between metric and imperial units
- **Tooltip Management**: Interactive information tooltips for all sensors
- **Event Handling**: Proper event propagation and click handling
- **Local Storage**: Persistent unit preferences
- **Error Handling**: Graceful degradation for API failures

**Key Functions:**
- `fetchWeatherData()` - Retrieves latest weather data from API
- `updateDisplay(data)` - Updates all dashboard elements with new data
- `toggleUnit(sensor)` - Switches between metric/imperial units
- `initializeCharts()` - Sets up Chart.js historical data charts
- `updateCharts(data)` - Updates charts with new data points
- `toggleTooltip(sensor)` - Shows/hides information tooltips

### `styles.css`
**Modern CSS Styling**
- Responsive grid layout that adapts to screen size
- Weather-themed color scheme with gradients
- Card-based design with hover effects
- Mobile-friendly responsive design
- Tooltip styling with dark theme
- Chart container styling and animations

### `date-fns.min.js`
**External Date Manipulation Library (v2.30.0)**
- Time-based chart axis formatting
- Date parsing and formatting utilities
- Chart.js date adapter integration
- Timezone handling for weather data timestamps

## Web Dashboard Features

### Real-time Weather Cards
1. **Temperature Card** - Air temperature with unit conversion (°C/°F)
2. **Humidity Card** - Relative humidity with heat index and comfort level descriptions
3. **Wind Card** - Speed, direction, and gust information with cardinal directions
4. **Rain Card** - Precipitation data with intensity descriptions and daily totals
5. **Pressure Card** - Barometric pressure with trend analysis and weather forecasting
6. **UV Index Card** - UV exposure levels with EPA color coding and risk categories
7. **Light Card** - Ambient light levels with illuminance context descriptions
8. **Forecast Card** - Weather predictions from Tempest API better_forecast endpoint

### Interactive Features
- **Unit Conversions**: Click any card to toggle units (persistent via localStorage)
- **Information Tooltips**: Click ℹ️ icons for detailed sensor information
- **Historical Charts**: Interactive Chart.js charts with zoom and pan capabilities
- **HomeKit Status**: Real-time display of HomeKit bridge and accessory status
- **Connection Status**: Live connection status to Tempest station

### Pressure Analysis System
The web server includes advanced pressure analysis capabilities:
- **Trend Detection**: Analyzes pressure changes over time (Rising/Falling/Stable)
- **Weather Forecasting**: Predicts weather conditions based on pressure patterns
- **Condition Assessment**: Real-time pressure condition evaluation (Normal/High/Low)

## Usage Examples

### Start Web Server
```go
import "tempest-homekit-go/pkg/web"

// Create and start web server
server := web.NewWebServer("8080")
go server.Start()
```

### Update Weather Data
```go
// Update server with new weather observation
server.UpdateWeather(weatherObservation)
```

### API Endpoints

#### Weather Data API
```
GET /api/weather
```
**Response:**
```json
{
  "temperature": 24.4,
  "humidity": 66.0,
  "windSpeed": 0.3,
  "windDirection": 241,
  "rainAccum": 0.0,
  "pressure": 979.7,
  "uv": 2,
  "illuminance": 15000,
  "lastUpdate": "2025-09-15T17:30:00Z"
}
```

#### Status API
```
GET /api/status
```
**Response:**
```json
{
  "connected": true,
  "lastUpdate": "2025-09-15T17:30:00Z",
  "uptime": "2h30m45s",
  "homekit": {
    "bridge": true,
    "accessories": 11,
    "pin": "00102003"
  }
}
```

## Frontend Architecture

### External JavaScript Architecture
All JavaScript code is externalized to `script.js` for clean separation of concerns:
- **HTML Templates**: Clean, semantic markup without inline scripts
- **CSS Styling**: Modern, responsive design with weather-themed colors
- **JavaScript Logic**: Complete interactivity in external file
- **Cache-Busting**: Static files served with timestamps to prevent caching issues

### Chart.js Integration
Historical data visualization using Chart.js v4.4.0:
- **Temperature Chart**: Line chart with temperature data over time
- **Humidity Chart**: Area chart showing humidity trends
- **Wind Chart**: Combination chart with speed and direction
- **Rain Chart**: Bar chart for precipitation data
- **Pressure Chart**: Line chart with trend analysis

### Responsive Design
The dashboard adapts to different screen sizes:
- **Desktop**: 3-column grid layout with full feature set
- **Tablet**: 2-column grid with optimized touch targets
- **Mobile**: Single-column layout with streamlined interface

## Static File Management

### Cache-Busting System
Static files are served with timestamps to prevent browser caching:
```
/pkg/web/static/script.js?t=1694808000
/pkg/web/static/styles.css?t=1694808000
```

### File Serving
- **Automatic Detection**: Server detects and serves static files
- **Content Types**: Proper MIME types for JS, CSS, and other assets
- **Error Handling**: Graceful fallbacks for missing files

## Testing

### Run Web Package Tests
```bash
go test ./pkg/web/... -v
go test ./pkg/web/... -cover
```

### HTTP Testing
Tests use `httptest` package for comprehensive endpoint testing:
- Mock HTTP requests and responses
- JSON parsing validation
- Error condition testing
- Static file serving verification

## Dependencies

### Go Dependencies
- **net/http**: HTTP server implementation
- **html/template**: HTML template rendering
- **encoding/json**: JSON API responses
- **path/filepath**: Static file serving
- **time**: Timestamp handling

### Frontend Dependencies
- **Chart.js**: Interactive charts and data visualization
- **date-fns**: Date manipulation and formatting
- **Vanilla JavaScript**: No frontend frameworks for simplicity

## Configuration

### Server Configuration
- **Port**: Configurable via `--web-port` flag (default: 8080)
- **Static Assets**: Served from `pkg/web/static/` directory
- **CORS**: Cross-origin requests allowed for API endpoints
- **Timeouts**: Configurable read/write timeouts for production use