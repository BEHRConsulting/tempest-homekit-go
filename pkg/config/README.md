# config/ Package

The `config` package handles all configuration management for the Tempest HomeKit Go application, including command-line flags, environment variables, and application settings.

## Files

### `config.go`
**Main Configuration Implementation**

**Core Functions:**
- `LoadConfig() *Config` - Loads configuration from flags and environment variables
- `ParseElevation(elevationStr string) (float64, error)` - Parses elevation strings with units (ft/m)
- Database path management and validation
- Default value handling

**Configuration Structure:**
```go
type Config struct {
    Token           string  // WeatherFlow API token
    StationName     string  // Tempest station name
    Pin             string  // HomeKit pairing PIN
    WebPort         string  // Web dashboard port
    LogLevel        string  // Logging verbosity (error/info/debug)
    Elevation       float64 // Station elevation in meters
    ReadHistory     bool    // Load historical weather data
    ClearDB         bool    // Reset HomeKit database
}
```

**Key Features:**
- **Elevation Parsing**: Supports both feet and meters (e.g., "1000ft", "300m")
- **Environment Variables**: Fallback to environment variables when flags aren't provided
- **Validation**: Validates required parameters and formats
- **Database Management**: Configures HomeKit database location

### `config_test.go`
**Comprehensive Unit Tests (66.4% Coverage)**

**Test Coverage:**
- Configuration loading from flags and environment variables
- Elevation parsing with various formats and edge cases
- Error handling for invalid inputs
- Default value validation
- Database path configuration

**Test Functions:**
- `TestLoadConfig()` - Configuration loading scenarios
- `TestParseElevation()` - Elevation parsing with various inputs
- `TestElevationEdgeCases()` - Error cases and validation
- `TestEnvironmentVariables()` - Environment variable precedence
- `TestDatabasePaths()` - Database configuration validation

## Usage Examples

### Basic Configuration Loading
```go
import "tempest-homekit-go/pkg/config"

// Load configuration (flags take precedence over env vars)
cfg := config.LoadConfig()

// Access configuration values
fmt.Printf("API Token: %s\n", cfg.Token)
fmt.Printf("Station: %s\n", cfg.StationName)
fmt.Printf("Elevation: %.2f meters\n", cfg.Elevation)
```

### Elevation Parsing
```go
// Parse elevation with units
elevation, err := config.ParseElevation("1200ft")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Elevation: %.2f meters\n", elevation) // Output: 365.76 meters

elevation, err = config.ParseElevation("300m")
fmt.Printf("Elevation: %.2f meters\n", elevation) // Output: 300.00 meters
```

## Command-Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--token` | string | "" | WeatherFlow API access token (required when using the WeatherFlow API as the data source) |
| `--station` | string | "Chino Hills" | Tempest station name |
| `--pin` | string | "00102003" | HomeKit pairing PIN |
| `--web-port` | string | "8080" | Web dashboard port |
| `--loglevel` | string | "error" | Logging level (error/warn/warning/info/debug) |
| `--logfilter` | string | "" | Filter log messages (case-insensitive substring match) |
| `--elevation` | string | "" | Station elevation (e.g., "1000ft", "300m") |
| `--read-history` | bool | false | Load historical weather data |
| `--cleardb` | bool | false | Reset HomeKit database |

## Environment Variables

| Variable | Corresponding Flag |
|----------|-------------------|
| `TEMPEST_TOKEN` | `--token` |
| `TEMPEST_STATION_NAME` | `--station` |
| `HOMEKIT_PIN` | `--pin` |
| `WEB_PORT` | `--web-port` |
| `LOG_LEVEL` | `--loglevel` |
| `LOG_FILTER` | `--logfilter` |

## Error Handling

The package provides robust error handling for:
- **Missing Required Parameters**: Validates that API token is provided
- **Invalid Elevation Formats**: Returns descriptive errors for malformed elevation strings
- **Unsupported Units**: Validates elevation units (only "ft" and "m" supported)
- **Invalid Numeric Values**: Handles non-numeric elevation values gracefully

## Testing

Run the configuration tests:
```bash
go test ./pkg/config/... -v
go test ./pkg/config/... -cover
```

The tests use table-driven testing patterns for comprehensive coverage of various configuration scenarios and edge cases.