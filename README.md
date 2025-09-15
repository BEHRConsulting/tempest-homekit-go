# Tempest HomeKit Go

A complete Go service application that monitors a WeatherFlow Tempest weather station and updates Apple HomeKit accessories with real-time weather data, enabling smart home automation based on temperature, humidity, wind speed, rain accumulation, and wind direction. Includes a modern web dashboard with interactive unit conversions and cross-platform deployment scripts.

## Authors

- **Kent** - Original project development and maintenance
- **GitHub Copilot** - AI-assisted development, code improvements, and documentation

## Features

- **Real-time Weather Monitoring**: Continuously polls WeatherFlow Tempest API for current weather observations every 60 seconds
- **Complete HomeKit Integration**: Updates 11 HomeKit sensors using modern `brutella/hap` library with custom services
- **Temperature Conversion Fix**: Custom weather services prevent HomeKit's automatic Celsius-to-Fahrenheit conversion
- **Accurate Weather Data**: Wind, rain, UV, and other sensors display exact values without unit conversion interference
- **Wind Direction Support**: Displays wind direction in cardinal format (N, NE, E, etc.) with degrees
- **Modern Web Dashboard**: Interactive web interface with real-time updates every 10 seconds and unit conversions
- **Cross-Platform Deployment**: Automated build and installation scripts for Linux, macOS, and Windows
- **Service Management**: Auto-start as system service with platform-specific managers (systemd, launchd, NSSM)
- **Modular Architecture**: Clean, maintainable code structure with separate packages
- **Enhanced Logging**: Multi-level logging (debug, info, error) with comprehensive sensor data
- **Command-line Interface**: Flexible configuration via flags and environment variables
- **Database Management**: Built-in `--cleardb` command to reset HomeKit pairing and rediscover device
- **Error Resilience**: Comprehensive error handling and graceful failure recovery

## Quick Start

### Prerequisites
- Go 1.24.2 or later
- WeatherFlow Tempest station with API access
- Apple device with HomeKit support

### Build and Run
```bash
git clone https://github.com/yourusername/tempest-homekit-go.git
cd tempest-homekit-go
go build
./tempest-homekit-go --token "your-api-token"
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
git clone https://github.com/yourusername/tempest-homekit-go.git
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
- Custom weather services with unique UUIDs to prevent temperature conversion issues

## Usage

### Basic Usage
```bash
./tempest-homekit-go --token "your-weatherflow-token"
```

### Configuration Options

#### Command-Line Flags
- `--token`: WeatherFlow API access token (required)
- `--station`: Tempest station name (default: "Chino Hills")
- `--pin`: HomeKit pairing PIN (default: "00102003")
- `--loglevel`: Logging level - debug, info, error (default: "error")
- `--web-port`: Web dashboard port (default: "8080")
- `--cleardb`: Clear HomeKit database and reset device pairing

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `WEB_PORT`: Web dashboard port

### Example with Full Configuration
```bash
./tempest-homekit-go \
  --token "your-api-token" \
  --station "Your Station Name" \
  --pin "12345678" \
  --web-port 8080 \
  --loglevel info
```

## HomeKit Setup

1. Start the application with your WeatherFlow API token
2. On your iOS device, open the Home app
3. Tap the "+" icon to add an accessory
4. Select "Don't have a code or can't scan?"
5. Choose the "Tempest Bridge"
6. Enter the PIN (default: 00102003)

The following sensors will appear as separate HomeKit accessories:
- **Temperature Sensor**: Air temperature in Celsius (uses standard HomeKit temperature characteristic)
- **Custom Wind Speed Sensor**: Wind speed in miles per hour (custom service prevents unit conversion)
- **Custom Wind Gust Sensor**: Wind gust speed in miles per hour (custom service)
- **Custom Wind Direction Sensor**: Wind direction in cardinal format with degrees (custom service)
- **Custom Humidity Sensor**: Relative humidity as percentage (custom service)
- **Custom Rain Sensor**: Rain accumulation in inches (custom service)
- **Custom UV Index Sensor**: UV index value (custom service)
- **Custom Lightning Count Sensor**: Lightning strike count (custom service)
- **Custom Lightning Distance Sensor**: Lightning strike distance (custom service)
- **Custom Precipitation Type Sensor**: Precipitation type indicator (custom service)
- **Custom Light Sensor**: Ambient light level in lux (custom service)

**Note**: Custom sensors use unique service UUIDs to prevent HomeKit's automatic temperature unit conversion, ensuring accurate display of weather data without unwanted Celsius-to-Fahrenheit conversion.

## Web Dashboard

Access the modern web dashboard at `http://localhost:8080` (or your configured port).

### Dashboard Features
- **Real-time Updates**: Weather data refreshes every 10 seconds
- **Interactive Unit Conversion**: Click any sensor card to toggle units:
  - 🌡️ **Temperature**: Celsius (°C) ↔ Fahrenheit (°F)
  - 🌬️ **Wind Speed**: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - 🌧️ **Rain**: Inches (in) ↔ Millimeters (mm)
- **Wind Direction Display**: Shows cardinal direction + degrees (e.g., "WSW (241°)")
- **Unit Persistence**: Preferences saved in browser localStorage
- **Modern Design**: Responsive interface with weather-themed styling
- **All Sensors**: Complete weather data display
- **HomeKit Status**: Bridge status, accessory count, and pairing PIN
- **Connection Status**: Real-time Tempest station connection status
- **Mobile Friendly**: Works perfectly on all devices

### API Endpoints
- `GET /`: Main dashboard with embedded HTML/CSS/JavaScript
- `GET /api/weather`: JSON weather data
- `GET /api/status`: Service and HomeKit status

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
│   │   └── client.go
│   ├── homekit/              # HomeKit accessory setup
│   │   └── setup.go
│   ├── web/                  # Web dashboard server
│   │   └── server.go
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
- 🚧 **Air Pressure**: Planned for future release

## Logging

### Log Levels
- **error**: Only errors and critical messages
- **info**: Basic operational messages + sensor data summary
- **debug**: Detailed sensor data + complete API JSON responses

### Example Log Output (Info Level)
```
2024-01-15 10:30:00 INFO Station found: Chino Hills (ID: 178915)
2024-01-15 10:30:00 INFO Weather update: Temp=72.5°F, Humidity=45%, Wind=3.2mph WSW, Rain=0.0in
2024-01-15 10:30:00 INFO HomeKit updated: 5 accessories
```

### Example Log Output (Debug Level)
```
2024-01-15 10:30:00 DEBUG API Response: {"status":{"status_code":0},"obs":[{"timestamp":1705312200,"air_temperature":72.5,"relative_humidity":45,"wind_avg":3.2,"wind_direction":247,"precip":0.0}]}
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

## Development

### Running Tests
```bash
go test ./...
```

### Building for Development
```bash
go build -o tempest-homekit-go
```

### Code Quality
- Comprehensive error handling and recovery
- Unit test coverage for all packages
- Modular design for maintainability
- Follows Go best practices and conventions

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
- ✅ Weather monitoring with 6 metrics (Temperature, Humidity, Wind Speed, Wind Direction, Rain, Light)
- ✅ Complete HomeKit integration with individual sensors
- ✅ Modern web dashboard with real-time updates
- ✅ Interactive unit conversions with persistence
- ✅ Cross-platform build and deployment
- ✅ Service management for all platforms
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error handling