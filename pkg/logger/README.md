# Logger Package

Centralized logging with level filtering and message filtering for the Tempest HomeKit application.

## Features

- **Level-based Filtering**: Control verbosity with `debug`, `info`, and `error` levels
- **Message Filtering**: Filter logs to show only messages containing specific strings
- **Case-insensitive Matching**: Filter string matching ignores case for easier debugging
- **Consistent Format**: All log messages include level prefix and timestamp

## Usage

### Basic Logging

```go
import "github.com/BEHRConsulting/tempest-homekit-go/pkg/logger"

// Set log level first
logger.SetLogLevel("debug")

// Log messages
logger.Debug("Detailed debug information: %s", debugData)
logger.Info("General information: %d items processed", count)
logger.Error("Error occurred: %v", err)
```

### Log Levels

- **`debug`**: Shows all messages (DEBUG, INFO, ERROR)
- **`info`**: Shows INFO and ERROR messages
- **`error`**: Shows only ERROR messages (default)

### Message Filtering

Filter logs to show only messages containing a specific string (case-insensitive):

```go
// Show only UDP-related messages
logger.SetLogFilter("udp")

// Now only messages containing "udp" (case-insensitive) will be output
logger.Info("UDP packet received")  // ✓ Shown
logger.Info("API call completed")    // ✗ Hidden
logger.Debug("Processing UDP data")  // ✓ Shown
```

### Command-Line Usage

```bash
# Enable debug logging
./tempest-homekit-go --loglevel debug

# Filter to show only UDP messages
./tempest-homekit-go --loglevel debug --logfilter "udp"

# Filter to show only forecast messages
./tempest-homekit-go --loglevel info --logfilter "forecast"

# Filter to show only parsed data
./tempest-homekit-go --loglevel debug --logfilter "parsed"
```

### Environment Variable

```bash
# Set log level via environment
export LOG_LEVEL=debug
export LOG_FILTER=udp
./tempest-homekit-go
```

## Common Filter Examples

| Filter | Shows |
|--------|-------|
| `"udp"` | All UDP broadcast messages |
| `"forecast"` | Forecast fetching and updates |
| `"parsed"` | Data parsing operations |
| `"observation"` | Weather observation processing |
| `"homekit"` | HomeKit accessory updates |
| `"web"` | Web server and dashboard activity |
| `"status"` | Station status updates |
| `"battery"` | Battery level information |

## Implementation Details

### Filter Logic

The filter performs a case-insensitive substring match on the formatted message:

```go
func shouldLog(message string) bool {
    if logFilter == "" {
        return true  // No filter, show all
    }
    return strings.Contains(strings.ToLower(message), logFilter)
}
```

### Performance

- Filter check is performed only for messages that pass the log level check
- No regular expressions - simple string matching for speed
- Message formatting only happens if the message will be displayed

## Examples

### Debugging UDP Stream Issues

```bash
./tempest-homekit-go --udp-stream --loglevel debug --logfilter "udp"
```

Output:
```
INFO: Log filter enabled: only messages containing 'udp' will be shown
INFO: UDP stream mode - will create UDP data source later
DEBUG: Parsed UDP message - Type: obs_st, Serial: ST-00163375
DEBUG: UDP Packet received (245 bytes): {"type":"obs_st",...}
INFO: UDP listener started on port 50222
```

### Monitoring Weather Data Updates

```bash
./tempest-homekit-go --loglevel debug --logfilter "observation"
```

### Tracking HomeKit Accessory Updates

```bash
./tempest-homekit-go --loglevel debug --logfilter "accessory"
```

## API Reference

### Functions

#### `SetLogLevel(level string)`
Sets the global log level. Valid values: `"debug"`, `"info"`, `"error"`

#### `SetLogFilter(filter string)`
Sets the global log filter string. Empty string disables filtering.

#### `Debug(format string, v ...interface{})`
Logs a debug message (only shown when log level is `debug`)

#### `Info(format string, v ...interface{})`
Logs an info message (shown when log level is `debug` or `info`)

#### `Error(format string, v ...interface{})`
Logs an error message (always shown)

## Testing

```go
import "testing"

func TestLogFilter(t *testing.T) {
    logger.SetLogLevel("debug")
    logger.SetLogFilter("test")
    
    // This would be shown
    logger.Info("Testing feature")
    
    // This would be hidden
    logger.Info("Production data")
}
```
