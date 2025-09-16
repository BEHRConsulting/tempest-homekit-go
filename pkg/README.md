# pkg/ Directory

This directory contains all the Go packages that comprise the Tempest HomeKit Go application. Each package has a specific responsibility in the overall architecture.

## Package Structure

### `config/`
**Configuration Management Package**
- Handles command-line flags, environment variables, and application configuration
- Provides elevation parsing (feet/meters), database paths, and service settings
- **Files:**
  - `config.go` - Main configuration structure and loading logic
  - `config_test.go` - Comprehensive unit tests with 66.4% coverage

### `homekit/`
**Apple HomeKit Integration Package**
- Manages HomeKit bridge and accessories setup
- Creates custom weather sensors with unique service UUIDs
- **Files:**
  - `modern_setup.go` - Modern HomeKit accessory setup using brutella/hap library
  - `custom_characteristics.go` - Custom weather sensor characteristics and service definitions

### `service/`
**Service Orchestration Package**
- Main service coordination and lifecycle management
- Handles graceful startup/shutdown and error recovery
- **Files:**
  - `service.go` - Service orchestration, polling loops, and signal handling
  - `service_test.go` - Unit tests for service functions (3.6% coverage)

### `weather/`
**WeatherFlow API Client Package**
- Communicates with WeatherFlow Tempest API
- Handles station discovery, data parsing, and API error management
- **Files:**
  - `client.go` - WeatherFlow API client implementation
  - `client_test.go` - Unit tests for API functions (16.2% coverage)

### `web/`
**Web Dashboard Package**
- HTTP server and web dashboard implementation
- Serves real-time weather data via REST API
- **Files:**
  - `server.go` - HTTP server, dashboard HTML, pressure analysis, and API endpoints
  - `server_test.go` - Unit tests for web server (50.5% coverage)
  - `static/` - Frontend assets (JavaScript, CSS, external libraries)

## Package Dependencies

```
main.go
├── pkg/config     (Configuration)
├── pkg/service    (Service Orchestration)
    ├── pkg/weather    (WeatherFlow API)
    ├── pkg/homekit    (HomeKit Integration)
    └── pkg/web        (Web Dashboard)
```

## Development

### Running Package Tests
```bash
# Test all packages
go test ./pkg/...

# Test specific package
go test ./pkg/config/...
go test ./pkg/weather/...
go test ./pkg/web/...
go test ./pkg/service/...

# Test with coverage
go test -cover ./pkg/...
```

### Package Documentation
View detailed package documentation using the GoDoc server:
```bash
./scripts/start-godoc.sh
# Visit: http://localhost:6060/pkg/tempest-homekit-go/pkg/
```

## Architecture Principles

- **Single Responsibility**: Each package has a focused purpose
- **Clean Interfaces**: Well-defined interfaces between packages
- **Error Handling**: Comprehensive error management throughout
- **Testing**: Unit tests for all critical functionality
- **Documentation**: Complete godoc documentation for all public APIs