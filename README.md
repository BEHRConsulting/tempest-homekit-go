# Tempest HomeKit Go: A Vibe Programming Case Study

A complete Go service application that monitors a WeatherFlow Tempest weather station and updates Apple HomeKit accessories with real-time weather data, enabling smart home automation based on weather conditions. This project serves as a comprehensive test case for **Vibe Programming** methodologies, demonstrating AI-assisted development techniques using modern Large Language Models.

**Version**: v1.4.0

## Research Methodology: Vibe Programming

### Definition of Vibe Programming

**Vibe Programming** represents a novel software development methodology that leverages the intuitive, context-aware capabilities of Large Language Models (LLMs) to enable rapid prototyping and iterative development through natural language interaction. This approach emphasizes:

- **Contextual Understanding**: LLMs maintain awareness of project architecture, requirements, and existing codebase
- **Iterative Refinement**: Continuous feedback loops between developer intent and AI-generated implementations
- **Natural Language Specifications**: Requirements expressed in conversational terms rather than formal specifications
- **Emergent Architecture**: System design evolving organically through AI-assisted exploration of possibilities
- **Real-time Problem Solving**: Immediate debugging and enhancement through conversational programming

### Technical Implementation Environment

**Primary Development Tools:**
- **IDE**: Visual Studio Code on macOS (Apple Silicon)
- **AI Assistants**: 
  - **Claude Sonnet 3.5** - Primary architectural design and complex problem resolution
  - **GitHub Copilot with Grok Code Fast 1 (preview)** - Code completion and rapid prototyping
- **Version Control**: Git with GitHub integration
- **Platform**: macOS development environment with cross-platform deployment

### Vibe Programming Validation Results

This project successfully demonstrates the efficacy of vibe programming techniques in producing production-ready software with:
- **Rapid Development Cycle**: Complete application developed through iterative AI assistance
- **Adaptive Problem Solving**: Real-time debugging and feature enhancement through conversational programming
- **Quality Assurance**: 78% test coverage achieved through AI-assisted test generation
- **Professional Standards**: Production-ready deployment with cross-platform service management

## Key v1.3.0 Enhancements

🚀 **Enhanced User Experience**:
- ✅ **Comprehensive Command Line Validation**: Detailed error messages with usage information for invalid arguments
- ✅ **Sensor Name Aliases**: Support for common sensor names (`temp`/`temperature`, `lux`/`light`, `uv`/`uvi`)
- ✅ **Earth-Realistic Elevation Validation**: Range validation (-430m to 8848m) with helpful geographic references
- ✅ **Complete Logging Compliance**: All log messages include proper level prefixes (DEBUG:, INFO:, WARNING:, ERROR:)
- ✅ **UV Value Precision**: UV sensor values now rounded to integers before HomeKit transmission
- ✅ **Improved Sensor Configuration**: Fixed "min" preset (temp,humidity,lux) and removed invalid "temp-only" preset
- ✅ **78% Test Coverage**: Comprehensive unit tests for command line validation and configuration parsing

## Important Sensor Notes

⚠️ **HomeKit Sensor Compliance**: Due to HomeKit's limited native sensor types, the **Pressure** and **UV Index** sensors use the standard HomeKit **Light Sensor** service for compliance. In the Home app, these will appear as "Light Sensor" with units showing as "lux" - **please ignore the "lux" unit** for these sensors as they represent atmospheric pressure (mb) and UV index values respectively. This is a HomeKit limitation, not an application issue.

🏠 **Web Console Only Mode**: This application can be run with HomeKit services completely disabled by using the `--disable-homekit` flag. In this mode, only the web dashboard will be available, providing a lightweight weather monitoring solution without HomeKit integration.

## Research Team

- **Kent** - Principal Investigator, Vibe Programming methodology implementation
- **Claude Sonnet 3.5** - Primary AI development partner for architectural design and complex problem resolution
- **GitHub Copilot (Grok Code Fast 1 preview)** - Secondary AI assistant for code completion and rapid prototyping

### Vibe Programming Methodology Validation

This project represents a controlled experiment in AI-assisted software development, demonstrating the practical application of conversational programming techniques in production software development.

## Features

- **Real-time Weather Monitoring**: Continuously polls WeatherFlow Tempest station data every 60 seconds
- **HomeKit Integration**: Individual HomeKit accessories for each weather sensor
- **Multiple Sensor Support**: Temperature, Humidity, Wind Speed, Wind Direction, Rain Accumulation, UV Index, Pressure, and Ambient Light
- **Modern Web Dashboard**: Interactive web interface with real-time updates, unit conversions, and professional styling
  - **Interactive Chart Pop-outs**: Click any chart to open in a resizable 80% screen window with complete historical data
  - **Professional Visualization**: Chart.js integration with draggable, resizable popup windows
  - **Full Dataset Display**: Pop-out windows show complete 1000+ point historical datasets with legends
- **Cross-platform Support**: Runs on macOS, Linux, and Windows with automated service installation
- **Flexible Configuration**: Command-line flags and environment variables for easy deployment

**Status**: ✅ **COMPLETE** - All planned features implemented and tested
- ✅ Weather monitoring with 11 HomeKit sensors (Temperature + 10 custom weather sensors)
- ✅ Complete HomeKit integration with compliance optimization
- ✅ Modern web dashboard with external JavaScript architecture
- ✅ TempestWX Device Status Scraping with `--use-web-status` flag:
  - ✅ Headless browser automation using Chrome/Chromium
  - ✅ 15-minute periodic updates with caching and fallback strategies
  - ✅ 12+ device status fields including battery voltage, uptimes, signal strength
  - ✅ Data source transparency with metadata and timestamps
  - ✅ Graceful fallback chain: Browser → HTTP → API → Fallback
- ✅ UV Index monitoring with NCBI reference data and EPA color coding
- ✅ Information tooltips system with standardized positioning
- ✅ HomeKit accessories status monitoring with enabled/disabled indicators
- ✅ Interactive unit conversions with localStorage persistence
- ✅ Cross-platform build and deployment with automated service management
- ✅ Professional styling and enhanced user experience
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error recovery

## Recent Major Improvements (September 2025)

### 🚀 Unified Data Architecture
- **Rain Totals Fixed**: Proper daily accumulation calculation (resolved 0.0 rain total bug)
- **Single Data Pipeline**: Eliminated complex branching between real and generated weather
- **Flexible Station URLs**: Support for custom weather endpoints with `--station-url`
- **Mock Tempest API**: Perfect API compatibility at `/api/generate-weather` endpoint
- **Clean Architecture**: Removed scattered special case handling throughout codebase

### 📊 Enhanced Visualization
- **Chart Improvements**: Fixed dataset rendering order for proper line visibility
- **Simplified Solar Charts**: Removed unnecessary average lines from light/UV charts
- **Better Tooltips**: All datasets now display properly in hover interactions
- **Data Accuracy**: Charts reflect actual weather data with proper timestamps

### 🔧 Technical Enhancements  
- **Backwards Compatibility**: `--use-generated-weather` still works seamlessly
- **Code Quality**: Single unified weather processing pipeline
- **Maintainability**: Much cleaner architecture without dual data paths

## Features

- **Real-time Weather Monitoring**: Continuously polls WeatherFlow Tempest station data every 60 seconds
- **HomeKit Integration**: Individual HomeKit accessories for each weather sensor
- **Multiple Sensor Support**: Temperature, Humidity, Wind Speed, Wind Direction, Rain Accumulation, UV Index, Pressure, and Ambient Light
- **Modern Web Dashboard**: Interactive web interface with real-time updates, unit conversions, and professional styling
  - **External JavaScript Architecture**: Clean separation of concerns with all JavaScript externalized to `script.js`
  - **Interactive Chart Pop-out System**: Advanced data visualization with expandable chart windows
    - **80% Screen Coverage**: Pop-out windows automatically sized to 80% of screen dimensions
    - **Resizable & Draggable**: Native browser window controls for optimal user experience
    - **Complete Historical Data**: Each pop-out displays full 1000+ point datasets with proper legends
    - **Professional Styling**: Gradient backgrounds with clean chart containers and interactive controls
    - **Multi-chart Support**: Temperature, humidity, wind, rain, pressure, light, and UV charts
  - **Pressure Analysis System**: Advanced pressure forecasting with trend analysis and weather predictions
  - **Interactive Info Icons**: Clickable info icons (ℹ️) with detailed tooltips for pressure calculations and sensor explanations
  - **Consistent Positioning**: All tooltips positioned with top-left corner aligned with bottom-right of info icons
  - **Rain Info Icon Fix**: Resolved JavaScript issue where unit updates removed the rain info icon
  - **Proper Event Handling**: Enhanced event propagation control to prevent unit toggle interference
  - **UV Index Display**: Complete UV exposure categories using NCBI reference data with EPA color coding
  - **Interactive Tooltips**: Information tooltips for all sensors with standardized positioning
  - **Accessories Status**: Real-time display of enabled/disabled sensor status in HomeKit bridge card
- **Cross-platform Support**: Runs on macOS, Linux, and Windows with automated service installation
- **TempestWX Device Status Scraping** (Optional):
  - **Headless Browser Integration**: Uses Chrome/Chromium to scrape detailed device status from TempestWX
  - **15-Minute Periodic Updates**: Background scraping with automatic caching
  - **Comprehensive Device Data**: Battery voltage, uptime, signal strength, firmware versions, serial numbers
  - **Multiple Fallback Layers**: Headless browser → HTTP scraping → API fallback for reliability
  - **Data Source Transparency**: Clear indication of data source (web-scraped, http-scraped, api, fallback)
  - **Enable with `--use-web-status` flag**: Optional enhancement for users who want detailed device monitoring
- **Flexible Configuration**: Command-line flags and environment variables for easy deployment
- **Enhanced Debug Logging**: Multi-level logging with emoji indicators, calculated values, API calls/responses, and comprehensive DOM debugging

## Quick Start

### Prerequisites
- Go 1.24.2 or later
- WeatherFlow Tempest station with API access
- Apple device with HomeKit support
- Google Chrome (optional, for detailed device status via `--use-web-status`)

### Build and Run
```bash
git clone https://github.com/BEHRConsulting/tempest-homekit-go.git
cd tempest-homekit-go
go build
./tempest-homekit-go --token "your-api-token"
```

### Test with Generated Weather
```bash
# Traditional approach
./tempest-homekit-go --use-generated-weather

# New flexible station URL approach
./tempest-homekit-go --station-url http://localhost:8080/api/generate-weather

# Using environment variable (equivalent to above)
STATION_URL=http://localhost:8080/api/generate-weather ./tempest-homekit-go

# With historical data preloading
./tempest-homekit-go --use-generated-weather --read-history
```

### Cross-Platform Build (All Platforms)
```bash
./scripts/build.sh
```

### Install as System Service
```bash
sudo ./scripts/install-service.sh --token "your-api-token"
```

## Installation

### Option 1: Build from Source
```bash
git clone https://github.com/BEHRConsulting/tempest-homekit-go.git
cd tempest-homekit-go
go mod tidy
go build -o tempest-homekit-go
```

### Option 2: Platform-Specific Build (Current Platform Only)
```bash
./scripts/build.sh
```
This builds only for your current platform (macOS binaries on macOS, Linux on Linux, etc.).

### Option 3: Cross-Platform Build (All Platforms)
```bash
./scripts/build-cross-platform.sh
```
This builds optimized binaries for Linux, macOS, and Windows from any platform.

### Option 3: Install as Service
For production deployment, install as a system service:
```bash
# Linux (systemd)
sudo ./scripts/install-service.sh --token "your-api-token"

# macOS (launchd)
sudo ./scripts/install-service.sh --token "your-api-token"

# Windows (NSSM)
./scripts/install-service.sh --token "your-api-token"
```

### Dependencies
- `github.com/brutella/hap` - Modern HomeKit Accessory Protocol implementation (v0.0.32)
- `github.com/chromedp/chromedp` - Headless browser automation for TempestWX status scraping
- Custom weather services with unique UUIDs to prevent temperature conversion issues

## Usage

### Basic Usage
```bash
./tempest-homekit-go --token "your-weatherflow-token"
```

### Configuration Options

#### Command-Line Flags (alphabetical order)
- `--cleardb`: Clear HomeKit database and reset device pairing
- `--disable-homekit`: Disable HomeKit services and run web console only
- `--elevation`: Station elevation in meters (default: auto-detect, valid range: -430m to 8848m)
- `--loglevel`: Logging level - debug, info, error (default: "error")
- `--pin`: HomeKit pairing PIN (default: "00102003")  
- `--sensors`: Sensors to enable - 'all', 'min' (temp,lux,humidity), or comma-delimited list with aliases supported:
  - **Temperature**: `temp` or `temperature`
  - **Light**: `lux` or `light`
  - **UV**: `uv` or `uvi`
  - **Other sensors**: `humidity`, `wind`, `rain`, `pressure`, `lightning`
  - (default: "temp,lux,humidity")
- `--station`: Tempest station name (default: "Chino Hills")
- `--station-url`: Custom station URL for weather data (e.g., `http://localhost:8080/api/generate-weather`). Overrides Tempest API
- `--token`: WeatherFlow API access token (required)*
- `--units`: Units system - imperial, metric, or sae (default: "imperial")
- `--units-pressure`: Pressure units - inHg or mb (default: "inHg")
- `--use-generated-weather`: Use simulated weather data for testing (automatically sets station-url)
- `--use-web-status`: Enable headless browser scraping of TempestWX status page every 15 minutes (requires Chrome)
- `--version`: Show version information and exit
- `--web-port`: Web dashboard port (default: "8080")

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `STATION_URL`: Custom station URL for weather data (overrides Tempest API)
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `SENSORS`: Sensors to enable (default: "temp,lux,humidity")
- `UNITS`: Units system - imperial, metric, or sae (default: "imperial")
- `UNITS_PRESSURE`: Pressure units - inHg or mb (default: "inHg")
- `WEB_PORT`: Web dashboard port

### Example with Full Configuration
```bash
./tempest-homekit-go \
  --token "your-api-token" \
  --station "Your Station Name" \
  --pin "12345678" \
  --web-port 8080 \
  --loglevel info \
  --sensors "temp,humidity,lux,uv,pressure" \
  --elevation 150 \
  --use-web-status
```

### Sensor Configuration Examples
```bash
# Using sensor aliases (recommended for readability)
./tempest-homekit-go --token "your-token" --sensors "temperature,light,uvi"

# Traditional sensor names (also supported)
./tempest-homekit-go --token "your-token" --sensors "temp,lux,uv"

# Mixed aliases and traditional names
./tempest-homekit-go --token "your-token" --sensors "temperature,humidity,light,wind"

# All available sensors
./tempest-homekit-go --token "your-token" --sensors "all"

# Minimal sensor set
./tempest-homekit-go --token "your-token" --sensors "min"
```

### Validation Examples
```bash
# Invalid elevation (too high) - shows helpful error message
./tempest-homekit-go --token "your-token" --elevation 10000
# Error: elevation must be between -430m and 8848m (Earth's surface range)

# Invalid sensor name - shows available options
./tempest-homekit-go --token "your-token" --sensors "invalid-sensor"
# Error: invalid sensor 'invalid-sensor'. Available: temp/temperature, lux/light, uv/uvi, humidity, wind, rain, pressure, lightning

# Missing required token - shows usage
./tempest-homekit-go --sensors "temp"
# Error: WeatherFlow API token is required. Use --token flag or TEMPEST_TOKEN environment variable
```

### Web Console Only (No HomeKit)
```bash
# Run web dashboard only without HomeKit services
./tempest-homekit-go \
  --token "your-api-token" \
  --disable-homekit \
  --web-port 8080 \
  --loglevel info
```

### TempestWX Device Status Scraping

Enable detailed device status monitoring with the `--use-web-status` flag:

```bash
# Basic usage with device status scraping
./tempest-homekit-go --token "your-token" --use-web-status

# With full configuration
./tempest-homekit-go --token "your-token" --use-web-status --loglevel debug
```

**Requirements:**
- Google Chrome or Chromium installed
- Internet access to https://tempestwx.com

**What it provides:**
- **Battery Status**: Real battery voltage (e.g., "2.69V") and condition (Good/Fair/Poor)
- **Device Uptime**: How long your Tempest device has been running
- **Hub Uptime**: How long your Tempest hub has been running  
- **Signal Strength**: Wi-Fi signal strength for hub, device signal strength
- **Firmware Versions**: Current firmware for both hub and device
- **Serial Numbers**: Hardware serial numbers for troubleshooting
- **Last Activity**: Timestamps of last status updates and observations

**Status API Response with Web Scraping:**
```json
{
  "stationStatus": {
    "batteryVoltage": "2.69V",
    "batteryStatus": "Good",
    "deviceUptime": "128d 6h 19m 29s",
    "hubUptime": "63d 15h 55m 1s",
    "hubWiFiSignal": "Strong (-42)",
    "deviceSignal": "Good (-65)",
    "hubSerialNumber": "HB-00168934",
    "deviceSerialNumber": "ST-00163375",
    "hubFirmware": "v329",
    "deviceFirmware": "v179",
    "dataSource": "web-scraped",
    "lastScraped": "2025-09-18T03:15:30Z",
    "scrapingEnabled": true
  }
}
```

**Without `--use-web-status` (default):**
Basic status with API-only data:
```json
{
  "stationStatus": {
    "batteryVoltage": "--",
    "dataSource": "api",
    "scrapingEnabled": false
  }
}
```

**How it works:**
1. **Headless Browser**: Launches Chrome to load the TempestWX status page
2. **JavaScript Execution**: Waits for JavaScript to populate the device status data
3. **Data Extraction**: Parses the loaded content to extract device information
4. **15-Minute Updates**: Automatically refreshes data every 15 minutes
5. **Graceful Fallbacks**: Falls back to HTTP scraping, then API-only if issues occur

## HomeKit Setup

1. Start the application with your WeatherFlow API token
2. On your iOS device, open the Home app
3. Tap the "+" icon to add an accessory
4. Select "Don't have a code or can't scan?"
5. Choose the "Tempest Bridge"
6. Enter the PIN (default: 00102003)

The following sensors will appear as separate HomeKit accessories:
- **Temperature Sensor**: Air temperature in Celsius (uses standard HomeKit temperature characteristic)
- **Humidity Sensor**: Relative humidity as percentage (uses standard HomeKit humidity characteristic)  
- **Light Sensor**: Ambient light level in lux (uses built-in HomeKit Light Sensor service)
- **Pressure Sensor**: Atmospheric pressure in mb (uses Light Sensor service for compliance - ignore "lux" unit label)
- **UV Index Sensor**: UV index value (uses Light Sensor service for compliance - ignore "lux" unit label)
- **Custom Wind Speed Sensor**: Wind speed in miles per hour (custom service prevents unit conversion)
- **Custom Wind Gust Sensor**: Wind gust speed in miles per hour (custom service)
- **Custom Wind Direction Sensor**: Wind direction in cardinal format with degrees (custom service)
- **Custom Rain Sensor**: Rain accumulation in inches (custom service)
- **Custom Lightning Count Sensor**: Lightning strike count (custom service)
- **Custom Lightning Distance Sensor**: Lightning strike distance (custom service)
- **Custom Precipitation Type Sensor**: Precipitation type indicator (custom service)

**Important**: The **Pressure** and **UV Index** sensors use HomeKit's standard Light Sensor service for maximum compatibility. In the Home app, they will appear as "Light Sensor" with "lux" units, but display the correct pressure (mb) and UV index values. Please ignore the "lux" unit label for these sensors - this is a HomeKit platform limitation, not an application issue.

⚠️ **HomeKit Compliance Warning**: As of Home.app v10.0, all sensors labeled as "(custom service)" above will return an "Out of Compliance" error when attempting to add the accessory to the Home app. Only the standard HomeKit services (Temperature, Humidity, Light, Pressure, UV Index) will successfully pair. This is due to Apple's stricter compliance enforcement in recent Home app versions.

## Web Dashboard

Access the modern web dashboard at `http://localhost:8080` (or your configured port).

### Dashboard Features
- **External JavaScript Architecture**: Clean separation with all ~800+ lines of JavaScript moved to external `script.js` file
- **Real-time Updates**: Weather data refreshes every 10 seconds with comprehensive error handling
- **Pressure Analysis System**: Advanced atmospheric pressure monitoring with:
  - **Trend Analysis**: Rising, Falling, or Stable pressure trends
  - **Weather Forecasting**: Predictions based on pressure patterns (Fair, Cloudy, Stormy)
  - **Interactive Info Icon**: Click the ℹ️ icon for detailed pressure calculation explanations
- **Interactive Unit Conversion**: Click any sensor card to toggle units:
  - 🌡️ **Temperature**: Celsius (°C) ↔ Fahrenheit (°F)
  - 🌬️ **Wind Speed**: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - 🌧️ **Rain**: Inches (in) ↔ Millimeters (mm)
  - 🌀 **Pressure**: Millibars (mb) ↔ Inches of Mercury (inHg)
- **UV Index Monitor**: 🌞 Complete UV exposure assessment with NCBI reference categories:
  - **Minimal (0-2)**: Low risk exposure with EPA green color coding
  - **Low (3-4)**: Moderate risk with yellow coding  
  - **Moderate (5-6)**: High risk with orange coding
  - **High (7-9)**: Very high risk with red coding
  - **Very High (10+)**: Extreme risk with violet coding
- **Enhanced Information System**: ℹ️ Detailed sensor tooltips with proper event propagation handling
- **Accessories Status**: Real-time HomeKit sensor status showing enabled/disabled state with priority sorting
- **Wind Direction Display**: Shows cardinal direction + degrees (e.g., "WSW (241°)")
- **Unit Persistence**: Preferences saved in browser localStorage
- **Modern Design**: Responsive interface with weather-themed styling and cache-busting script loading
- **All Sensors**: Complete weather data display with comprehensive DOM debugging
- **HomeKit Status**: Bridge status, accessory count, and pairing PIN
- **Connection Status**: Real-time Tempest station connection status
- **Mobile Friendly**: Works perfectly on all devices with enhanced event listener management

### API Endpoints
- `GET /`: Main dashboard HTML with external JavaScript
- `GET /pkg/web/static/script.js`: External JavaScript file with cache-busting timestamps
- `GET /api/weather`: JSON weather data with pressure analysis
- `GET /api/status`: Service and HomeKit status with optional TempestWX device status

## Architecture

```
tempest-homekit-go/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── scripts/
│   ├── build.sh              # Platform-specific build script
│   ├── build-cross-platform.sh # Cross-platform build script
│   ├── install-service.sh    # Service installation script
│   ├── remove-service.sh     # Service removal script
│   └── README.md             # Scripts documentation
├── pkg/
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── weather/              # WeatherFlow API client
│   │   ├── client.go         # API client and TempestWX scraping
│   │   └── status_manager.go # Periodic status scraping manager
│   ├── homekit/              # HomeKit accessory setup
│   │   ├── modern_setup.go   # Modern HAP library implementation
│   │   └── custom_characteristics.go # Custom weather characteristics
│   ├── web/                  # Web dashboard server
│   │   ├── server.go         # HTTP server with static file serving
│   │   └── static/           # Static web assets
│   │       ├── script.js     # External JavaScript (~800+ lines)
│   │       ├── styles.css    # CSS styling
│   │       └── date-fns.min.js # Date manipulation library
│   └── service/              # Main service orchestration
│       └── service.go
└── README.md
```

## API Integration

### WeatherFlow Tempest API
- **Stations Endpoint**: `GET /swd/rest/stations?token={token}`
- **Observations Endpoint**: `GET /swd/rest/observations/station/{station_id}?token={token}`

### Supported Weather Metrics
- ✅ **Air Temperature**: In Fahrenheit/Celsius
- ✅ **Relative Humidity**: As percentage
- ✅ **Wind Speed**: Average wind speed in mph/kph
- ✅ **Wind Direction**: Degrees with cardinal conversion
- ✅ **Rain Accumulation**: Total precipitation in inches/mm
- ✅ **Air Pressure**: Atmospheric pressure in mb/inHg
- ✅ **UV Index**: UV exposure level (0-15)
- ✅ **Ambient Light**: Illuminance in lux

## Logging

### Log Levels
- **error**: Only errors and critical messages
- **info**: Basic operational messages + sensor data summary
- **debug**: Detailed sensor data + complete API JSON responses

### Example Log Output (Info Level)
```
2025-09-21 10:30:00 Starting service with config: WebPort=8080, LogLevel=info
2025-09-21 10:30:00 Starting Tempest HomeKit service...
2025-09-21 10:30:00 Found station: Chino Hills (ID: 178915)
2025-09-21 10:30:00 INFO: HomeKit server started successfully with PIN: 00102003
2025-09-21 10:30:00 INFO: Starting web dashboard on port 8080
2025-09-21 10:30:00 Starting web server on port 8080
2025-09-21 10:30:00 INFO: Successfully read weather data from Tempest API - Station: Chino Hills
2025-09-21 10:30:00 INFO: Sensor data - Temp: 22.7°C, Humidity: 77%, Wind: 0.3 mph (238°), Rain: 0.000 in, Light: 1 lux
```

### Example Log Output (Debug Level)
```
2025-09-21 10:30:00 service.go:25: Starting Tempest HomeKit service...
2025-09-21 10:30:00 service.go:29: DEBUG: Fetching stations from WeatherFlow API
2025-09-21 10:30:00 modern_setup.go:39: DEBUG: Creating new weather system with hap library
2025-09-21 10:30:00 modern_setup.go:89: DEBUG: Created temperature sensor accessory
2025-09-21 10:30:00 modern_setup.go:169: DEBUG: Created UV Index sensor accessory using light sensor service with UV range
2025-09-21 10:30:00 service.go:284: DEBUG: HomeKit - UV Index: 0
2025-09-21 10:30:00 service.go:304: DEBUG: Updating UV Index: 0.000
```

### Example Log Output (Error Level - Default)
```
2025-09-21 10:30:00 Starting service with config: WebPort=8080, LogLevel=error
2025-09-21 10:30:00 Starting Tempest HomeKit service...
2025-09-21 10:30:00 Found station: Chino Hills (ID: 178915)
2025-09-21 10:30:00 Starting web server on port 8080
```

## Service Management

### Linux (systemd)
```bash
# Install
sudo ./scripts/install-service.sh --token "your-token"

# Check status
sudo systemctl status tempest-homekit-go

# View logs
sudo journalctl -u tempest-homekit-go -f

# Remove
sudo ./scripts/remove-service.sh
```

### macOS (launchd)
```bash
# Install
sudo ./scripts/install-service.sh --token "your-token"

# Check status
sudo launchctl list | grep tempest

# View logs
log show --predicate 'process == "tempest-homekit-go"' --last 1h

# Remove
sudo ./scripts/remove-service.sh
```

### Windows (NSSM)
```bash
# Install
./scripts/install-service.sh --token "your-token"

# Check status
sc query tempest-homekit-go

# View logs (via Event Viewer)
# Remove
./scripts/remove-service.sh
```

## Configuration

### WeatherFlow API Token
1. Visit [tempestwx.com](https://tempestwx.com)
2. Go to Settings → Data Authorizations
3. Create a new personal access token
4. Use with `--token` flag or `TEMPEST_TOKEN` environment variable

### Station Discovery
The application automatically finds your station by name. Ensure your station name in WeatherFlow matches the `--station` parameter.

## Troubleshooting

### HomeKit Re-pairing (Database Reset)

When you make changes to HomeKit accessories (such as modifying sensor types, names, or configurations), you may need to reset the HomeKit database and re-pair the bridge with your Home app. This ensures the changes take effect properly.

#### Using the Built-in --cleardb Command (Recommended)

The easiest way to reset HomeKit pairing is using the built-in `--cleardb` command:

```bash
# Stop the current service if running
pkill -f tempest-homekit-go

# Clear the database and reset pairing
./tempest-homekit-go --cleardb

# Restart the service normally
./tempest-homekit-go --token "your-api-token"
```

#### Manual Database Reset

If you prefer to do it manually:

1. **Stop the Application**
   ```bash
   # If running as a service
   sudo systemctl stop tempest-homekit-go  # Linux
   sudo launchctl stop tempest-homekit-go  # macOS
   sc stop tempest-homekit-go              # Windows
   
   # Or kill the process directly
   pkill -f tempest-homekit-go
   ```

2. **Delete the HomeKit Database**
   ```bash
   # Navigate to the application directory
   cd /path/to/tempest-homekit-go
   
   # Remove the database directory (this contains all pairing information)
   rm -rf ./db/
   
   # Verify the directory is empty
   ls -la ./db/
   ```

3. **Restart the Application**
   ```bash
   # Start the application again
   ./tempest-homekit-go --token "your-api-token"
   
   # Or restart the service
   sudo systemctl start tempest-homekit-go  # Linux
   sudo launchctl start tempest-homekit-go  # macOS
   sc start tempest-homekit-go              # Windows
   ```

4. **Re-pair in Home App**
   - Open the Home app on your iOS device
   - The "Tempest HomeKit Bridge" should appear as a new, unpaired accessory
   - Tap the "+" icon to add an accessory
   - Select "Don't have a code or can't scan?"
   - Choose the "Tempest HomeKit Bridge"
   - Enter the PIN (default: `00102003`)

5. **Verify the Changes**
   - Check that all accessories appear correctly
   - Confirm sensor types and names are as expected
   - Test that sensors are no longer grouped incorrectly

#### Important Notes:
- **Data Loss**: This will remove all HomeKit pairing information and automation rules
- **Re-setup Required**: You'll need to re-add any scenes, automations, or accessory groupings
- **Safe Operation**: The weather data collection continues normally; only HomeKit pairing is affected
- **Backup First**: Consider noting any important automation rules before resetting

#### Alternative: Clear Specific Database Files
If you want to be more selective, you can remove specific database files instead of the entire directory:
```bash
# Remove only pairing information (keeps other HomeKit data)
rm -f ./db/pairings.json

# Remove accessory cache (forces rediscovery)
rm -f ./db/accessories.json
```

### Common Issues
- **"Station not found"**: Verify station name matches exactly (case-sensitive)
- **"API request failed"**: Check internet connection and API token validity
- **HomeKit pairing fails**: Ensure PIN is correct and no other devices are pairing
- **Web dashboard not loading**: Check if port 8080 is available
- **Sensors showing wrong values/types**: Reset HomeKit database and re-pair (see above)

### Debug Mode
Enable detailed logging for troubleshooting:
```bash
./tempest-homekit-go --loglevel debug --token "your-token"
```

### Service Issues
```bash
# Check service status
./scripts/install-service.sh --status

# Restart service
./scripts/remove-service.sh
./scripts/install-service.sh --token "your-token"
```

## Recent Enhancements

### Tooltip Positioning & User Experience Improvements (Latest)
- **Consistent Tooltip Positioning**: All information tooltips now open with their top-left corner aligned with the bottom-right of their respective info icons
- **Rain Info Icon Resolution**: Fixed JavaScript issue where `updateUnits()` function was removing the rain info icon during unit conversions
- **Enhanced Event Handling**: Implemented proper event propagation control with `stopPropagation()` to prevent interference between info icon clicks and unit toggles
- **Humidity Description Addition**: Added visible humidity comfort level descriptions below units, matching the lux card pattern
- **Context Container Architecture**: Added proper `position: relative` containers for all tooltips to ensure consistent positioning behavior

### JavaScript Architecture Modernization
- **Complete Separation of Concerns**: Moved all ~800+ lines of JavaScript from HTML template to external `script.js`
- **Cache-Busting File Serving**: Static files served with timestamps to prevent browser caching issues
- **Enhanced Event Management**: MutationObserver-based event listener attachment with retry mechanisms
- **Comprehensive DOM Debugging**: Advanced element detection and interaction logging
- **Improved Maintainability**: Clean HTML templates with external asset references

### Pressure Analysis System
- **Advanced Forecasting**: Server-side pressure trend analysis with weather predictions
- **Interactive Info Icons**: Clickable ℹ️ icons with detailed calculation explanations
- **Event Propagation Control**: Proper handling of nested click events to prevent unit toggle interference
- **Real-time Calculations**: Live pressure condition assessment (Normal, High, Low) with trend indicators

### Enhanced Debug Logging
- **Multi-Level System**: DEBUG, INFO, WARN, ERROR levels with emoji indicators (🐛, ℹ️, ⚠️, ❌)
- **Structured Data Output**: API calls/responses, calculated values, and sensor updates
- **DOM Inspection Tools**: Comprehensive element detection and HTML content analysis
- **Event Listener Monitoring**: Detailed tracking of event attachment and retry attempts

### UV Index Monitoring
- **Complete UV Exposure Assessment**: Professional UV Index display with NCBI reference categories
- **EPA Color Coding**: Visual risk indicators from green (minimal) to violet (extreme)
- **Real-time Updates**: Live UV Index monitoring with automatic risk category assessment
- **Educational Information**: Tooltip with detailed exposure risk information

### Information Tooltips System
- **Standardized Positioning**: All information tooltips consistently aligned for optimal visibility
- **Rich Content**: Detailed sensor information including measurement ranges and units
- **Professional Design**: Clean tooltip styling with proper contrast and readability
- **Event Handling**: Proper click event management with stopPropagation for nested elements

### HomeKit Accessories Status
- **Real-time Status Monitoring**: Live display of enabled/disabled sensor status in web dashboard
- **Priority Sorting**: Active sensors automatically sorted to the top of the accessories list
- **Clear Visual Indicators**: Distinct styling for enabled vs disabled accessories
- **Configuration Transparency**: Shows exactly which sensors are currently being provided to HomeKit

### Enhanced User Experience
- **Consistent Design Language**: Unified styling across all dashboard components
- **Improved Accessibility**: Better contrast ratios and screen reader support  
- **Responsive Layout**: Enhanced mobile experience with optimized touch targets
- **Performance Optimizations**: Faster dashboard loading and more efficient updates

## Development

### GoDoc Server
Browse the complete Go documentation and API references locally:
```bash
# Start GoDoc server on port 6060 (opens browser automatically)
./scripts/start-godoc.sh

# Start on custom port without opening browser
./scripts/start-godoc.sh --port 8080 --no-browser

# View help
./scripts/start-godoc.sh --help
```

Then visit `http://localhost:6060` to browse:
- Package documentation for all modules (`pkg/config`, `pkg/weather`, etc.)
- Function and type definitions with examples
- Cross-referenced source code
- Standard library documentation

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run verbose tests
go test -v ./...

# Run specific package tests
go test ./pkg/config/...
go test ./pkg/weather/...
go test ./pkg/web/...
go test ./pkg/service/...
```

### Test Coverage Overview
- **pkg/config**: 66.4% coverage - Configuration management, elevation parsing, database operations
- **pkg/weather**: 16.2% coverage - WeatherFlow API client, data parsing utilities, station discovery
- **pkg/web**: 50.5% coverage - HTTP server, pressure analysis, status endpoints
- **pkg/service**: 3.6% coverage - Service orchestration, logging, environmental detection

### Testing Architecture
The project includes comprehensive unit tests covering:
- **Configuration Management**: Flag parsing, environment variables, elevation parsing (feet/meters)
- **Weather Client**: Station discovery, device ID extraction, JSON parsing helpers, time filtering
- **Web Server**: HTTP endpoints, pressure trend analysis, history loading progress
- **Service Functions**: Log level management, night time detection based on illuminance

### Building for Development
```bash
go build -o tempest-homekit-go
```

### Code Quality
- Comprehensive error handling and recovery
- Unit test coverage for all packages with table-driven tests
- Modular design for maintainability
- Follows Go best practices and conventions
- HTTP testing with `httptest.ResponseRecorder`
- Mock data creation for realistic test scenarios

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- **WeatherFlow** for the Tempest weather station and API
- **Apple** for the HomeKit platform
- **hc library** for HomeKit Go implementation
- **Community** for feedback and contributions

## References

This project was developed using various technologies, libraries, and tools. Below is a comprehensive list of key components and resources that contributed to the development:

### Core Technologies
- **Go Programming Language** (v1.24.2+) - Primary programming language
- **HomeKit Accessory Protocol** - Apple's smart home communication protocol
- **WeatherFlow Tempest API** - Weather data source and API integration

### Go Libraries and Dependencies
- **`github.com/brutella/hap`** - HomeKit Accessory Protocol implementation for Go
- **Standard Library Packages**:
  - `net/http` - Web server implementation
  - `encoding/json` - JSON data handling
  - `sync` - Concurrent programming primitives
  - `time` - Time and date operations
  - `log` - Logging functionality
  - `os` - Operating system interface
  - `flag` - Command-line flag parsing

### Web Technologies (Embedded Dashboard)
- **HTML5** - Dashboard structure and markup
- **CSS3** - Responsive styling and animations
- **JavaScript (ES6+)** - Interactive functionality and real-time updates
- **Chart.js** (v4.4.0) - Interactive charts and data visualization
- **date-fns** (v2.30.0) - Date and time manipulation in JavaScript
- **Chart.js Date-Fns Adapter** (v2.0.1) - Time-based chart integration

### Development Tools and AI Assistance
- **GitHub Copilot** - AI-powered code suggestions and development assistance
- **Visual Studio Code** - Primary development environment
- **Go Modules** - Dependency management
- **Git** - Version control system

### Platform-Specific Tools
- **systemd** (Linux) - Service management
- **launchd** (macOS) - Service management
- **NSSM** (Windows) - Non-Sucking Service Manager for Windows services

### Build and Deployment
- **Cross-compilation** - Go's built-in cross-platform compilation
- **Shell scripting** - Bash scripts for automated builds and deployment
- **Platform detection** - Runtime OS and architecture detection

### External Resources and Documentation
- **WeatherFlow API Documentation** - Weather data integration reference
- **Apple HomeKit Developer Documentation** - HomeKit protocol implementation guide
- **Go Documentation** - Standard library and language reference
- **MDN Web Docs** - JavaScript, HTML, and CSS reference

### Development Practices
- **Test-Driven Development** - Unit testing approach
- **Modular Architecture** - Clean code organization
- **Error Handling** - Comprehensive error management
- **Logging** - Multi-level logging system
- **Configuration Management** - Flexible configuration via flags and environment variables

---

**Status**: ✅ **COMPLETE** - All planned features implemented and tested
- ✅ Weather monitoring with 11 HomeKit sensors (Temperature + 10 custom weather sensors)
- ✅ Complete HomeKit integration with compliance optimization
- ✅ Modern web dashboard with real-time updates and interactive features
- ✅ UV Index monitoring with NCBI reference data and EPA color coding
- ✅ Information tooltips system with standardized positioning
- ✅ HomeKit accessories status monitoring with enabled/disabled indicators
- ✅ Interactive unit conversions with localStorage persistence
- ✅ Cross-platform build and deployment with automated service management
- ✅ Professional styling and enhanced user experience
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error recovery
- ✅ Weather monitoring with 6 metrics (Temperature, Humidity, Wind Speed, Wind Direction, Rain, Light)
- ✅ Complete HomeKit integration with individual sensors
- ✅ Modern web dashboard with real-time updates
- ✅ Interactive unit conversions with persistence
- ✅ Cross-platform build and deployment
- ✅ Service management for all platforms
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error handling