# Tempest HomeKit Go

A Go service applic### Command-## Tempest HomeKit Go

A Go service application that monitors a WeatherFlow Tempest weather station and updates Apple HomeKit accessories with real-time weather data, enabling automation based on temperature, humidity, wind speed, rain accumulation, and wind direction.

## Features

- **Real-time Weather Monitoring**: Continuously polls WeatherFlow Tempest API for current weather observations
- **HomeKit Integration**: Updates HomeKit temperature, humidity, wind speed, and rain sensors
- **Wind Direction Support**: Displays wind direction in cardinal format (N, NE, E, etc.) with degrees
- **Modern Web Dashboard**: Interactive web interface with real-time updates and unit conversions
- **Modular Architecture**: Clean, maintainable code structure with separate packages
- **Configurable Logging**: Adjustable log levels (debug, info, error)
- **Command-line Interface**: Flexible configuration via flags and environment variables
- **Error Resilience**: Comprehensive error handling and graceful failure recovery

## Installation

### Prerequisites
- Go 1.19 or later
- WeatherFlow Tempest station with API access
- Apple device with HomeKit support

### Build from Source
```bash
git clone https://github.com/yourusername/tempest-homekit-go.git
cd tempest-homekit-go
go build
```

### Dependencies
The application uses the following Go modules:
- `github.com/brutella/hc` - HomeKit Accessory Protocol implementation

## Usage

### Basic Usage
```bash
./tempest-homekit-go
```

### Configuration
Configure the application using command-line flags or environment variables:

#### Command-Line Flags
- `--token`: WeatherFlow API access token
- `--station`: Tempest station name (default: "Chino Hills")
- `--pin`: HomeKit PIN (default: "00102003")
- `--loglevel`: Logging level - debug, info, error (default: "error")
- `--web-port`: Web dashboard port (default: "8080")

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `WEB_PORT`: Web dashboard port

### Example
```bash
./tempest-homekit-go --token "your-api-token" --station "your-station-name" --web-port 8080 --loglevel info
```

## HomeKit Setup

1. Run the application
2. On your iOS device, open the Home app
3. Tap the "+" icon to add an accessory
4. Select "Don't have a code or can't scan?"
5. Choose the "Tempest HomeKit Bridge"
6. Enter the PIN (default: 00102003)

The temperature, humidity, wind speed, rain accumulation, and wind direction sensors will appear as separate accessories.

## Web Dashboard

The application includes a modern web dashboard for monitoring weather data in real-time with interactive unit conversions:

### Accessing the Dashboard
1. Start the application with `./tempest-homekit-go --web-port 8080`
2. Open your web browser and navigate to `http://localhost:8080`
3. View real-time weather data with an elegant, responsive interface

### Dashboard Features
- **Real-time Updates**: Weather data updates every 10 seconds
- **Interactive Unit Conversion**: Click temperature, wind, or rain cards to toggle between units:
  - Temperature: Celsius (°C) ↔ Fahrenheit (°F)
  - Wind Speed: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - Rain: Inches (in) ↔ Millimeters (mm)
- **Wind Direction Display**: Shows wind direction in cardinal format with degrees (e.g., "WSW (241°)")
- **Unit Persistence**: Your unit preferences are saved in browser localStorage
- **Modern Design**: Clean, responsive interface with weather-themed styling
- **All Sensors**: Displays temperature, humidity, wind speed, rain accumulation, and wind direction
- **HomeKit Status**: Shows HomeKit bridge status, accessory count, and PIN
- **Connection Status**: Real-time connection status to the Tempest station
- **Mobile Friendly**: Responsive design that works on all devices

### API Endpoints
- `GET /`: Main dashboard page
- `GET /api/weather`: JSON API for current weather data
- `GET /api/status`: JSON API for service and HomeKit status

## Architecture

```
tempest-homekit-go/
├── pkg/
│   ├── config/     # Configuration management
│   ├── weather/    # WeatherFlow API client
│   ├── homekit/    # HomeKit accessory setup
│   ├── web/        # Web dashboard server
│   └── service/    # Main service orchestration
├── main.go         # Application entry point
├── go.mod          # Go module definition
├── go.sum          # Dependency checksums
└── README.md
```

## API Integration

The application integrates with:
- **WeatherFlow Tempest API**: REST API for weather station data
- **Apple HomeKit**: HAP (HomeKit Accessory Protocol) for smart home integration

### Supported Weather Metrics
- ✅ Air Temperature
- ✅ Relative Humidity
- ✅ Wind Speed (average)
- ✅ Wind Direction (cardinal + degrees)
- ✅ Rain Accumulation
- 🚧 Air Pressure (planned)

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build
```

### Code Quality
- Follows Go best practices
- Comprehensive error handling
- Unit test coverage
- Modular design for maintainability

## Configuration

### WeatherFlow API Token
Obtain your personal access token from the WeatherFlow web app:
1. Visit tempestwx.com
2. Go to Settings → Data Authorizations
3. Create a new token
4. Use the token with the `--token` flag or `TEMPEST_TOKEN` environment variable

### Station Name
The application automatically discovers your station by name. Ensure your station is named appropriately in the WeatherFlow app.

## Troubleshooting

### Common Issues
- **"Station not found"**: Verify the station name matches exactly
- **"API request failed"**: Check your internet connection and API token
- **HomeKit pairing fails**: Ensure the PIN is correct and no other devices are trying to pair

### Logs
Increase log verbosity for debugging:
```bash
./tempest-homekit-go --loglevel debug
```

When debug logging is enabled, the application will log detailed weather data including temperature, humidity, wind speed, wind direction, and rain accumulation with each API poll.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- WeatherFlow for the Tempest weather station and API
- Apple for the HomeKit platform
- The hc library for HomeKit Go implementationFlags
- `--token`: WeatherFlow API access token
- `--station`: Tempest station name (default: "Chino Hills")
- `--pin`: HomeKit PIN (default: "00102003")
- `--loglevel`: Logging level - debug, info, error (default: "error")
- `--web-port`: Web dashboard port (default: "8080")

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `WEB_PORT`: Web dashboard portmonitors a WeatherFlow Tempest weather station and updates Apple HomeKit accessories with real-time weather data, enabling automation based on temperature, humidity, and other weather conditions.

## Features

- **Real-time Weather Monitoring**: Continuously polls WeatherFlow Tempest API for current weather observations
- **HomeKit Integration**: Updates HomeKit temperature and humidity sensors
- **Modular Architecture**: Clean, maintainable code structure with separate packages
- **Configurable Logging**: Adjustable log levels (debug, info, error)
- **Command-line Interface**: Flexible configuration via flags and environment variables
- **Error Resilience**: Comprehensive error handling and graceful failure recovery

## Installation

### Prerequisites
- Go 1.19 or later
- WeatherFlow Tempest station with API access
- Apple device with HomeKit support

### Build from Source
```bash
git clone https://github.com/yourusername/tempest-homekit-go.git
cd tempest-homekit-go
go build
```

### Dependencies
The application uses the following Go modules:
- `github.com/brutella/hc` - HomeKit Accessory Protocol implementation

## Usage

### Basic Usage
```bash
./tempest-homekit-go
```

### Configuration
Configure the application using command-line flags or environment variables:

#### Command-line Flags
- `--token`: WeatherFlow API access token
- `--station`: Tempest station name (default: "Chino Hills")
- `--pin`: HomeKit PIN (default: "00102003")
- `--loglevel`: Logging level - debug, info, error (default: "error")

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `WEB_PORT`: Web dashboard port

### Example
```bash
./tempest-homekit-go --token "your-api-token" --station "your-station-name" --web-port 8080 --loglevel info
```

## HomeKit Setup

1. Run the application
2. On your iOS device, open the Home app
3. Tap the "+" icon to add an accessory
4. Select "Don't have a code or can't scan?"
5. Choose the "Tempest HomeKit Bridge"
6. Enter the PIN (default: 00102003)

The temperature, humidity, wind speed, and rain accumulation sensors will appear as separate accessories.

## Web Dashboard

The application includes a modern web dashboard for monitoring weather data in real-time with interactive unit conversions:

### Accessing the Dashboard
1. Start the application with `./tempest-homekit-go --web-port 8080`
2. Open your web browser and navigate to `http://localhost:8080`
3. View real-time weather data with an elegant, responsive interface

### Dashboard Features
- **Real-time Updates**: Weather data updates every 10 seconds
- **Interactive Unit Conversion**: Click temperature, wind, or rain cards to toggle between units:
  - Temperature: Celsius (°C) ↔ Fahrenheit (°F)
  - Wind Speed: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - Rain: Inches (in) ↔ Millimeters (mm)
- **Unit Persistence**: Your unit preferences are saved in browser localStorage
- **Modern Design**: Clean, responsive interface with weather-themed styling
- **All Sensors**: Displays temperature, humidity, wind speed, and rain accumulation
- **HomeKit Status**: Shows HomeKit bridge status, accessory count, and PIN
- **Connection Status**: Real-time connection status to the Tempest station
- **Mobile Friendly**: Responsive design that works on all devices

### API Endpoints
- `GET /`: Main dashboard page
- `GET /api/weather`: JSON API for current weather data
- `GET /api/status`: JSON API for service and HomeKit status

## Architecture

```
tempest-homekit-go/
├── pkg/
│   ├── config/     # Configuration management
│   ├── weather/    # WeatherFlow API client
│   ├── homekit/    # HomeKit accessory setup
│   └── service/    # Main service orchestration
├── main.go         # Application entry point
└── README.md
```

## API Integration

The application integrates with:
- **WeatherFlow Tempest API**: REST API for weather station data
- **Apple HomeKit**: HAP (HomeKit Accessory Protocol) for smart home integration

### Supported Weather Metrics
- ✅ Air Temperature
- ✅ Relative Humidity
- ✅ Wind Speed (average)
- ✅ Rain Accumulation
- 🚧 Air Pressure (planned)
- 🚧 Wind Direction (planned)

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build
```

### Code Quality
- Follows Go best practices
- Comprehensive error handling
- Unit test coverage
- Modular design for maintainability

## Configuration

### WeatherFlow API Token
Obtain your personal access token from the WeatherFlow web app:
1. Visit tempestwx.com
2. Go to Settings → Data Authorizations
3. Create a new token
4. Use the token with the `--token` flag or `TEMPEST_TOKEN` environment variable

### Station Name
The application automatically discovers your station by name. Ensure your station is named appropriately in the WeatherFlow app.

## Troubleshooting

### Common Issues
- **"Station not found"**: Verify the station name matches exactly
- **"API request failed"**: Check your internet connection and API token
- **HomeKit pairing fails**: Ensure the PIN is correct and no other devices are trying to pair

### Logs
Increase log verbosity for debugging:
```bash
./tempest-homekit-go --loglevel debug
```

When debug logging is enabled, the application will log detailed weather data including temperature, humidity, wind speed, and rain accumulation with each API poll.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- WeatherFlow for the Tempest weather station and API
- Apple for the HomeKit platform
- The hc library for HomeKit Go implementation