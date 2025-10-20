# Logger Package

Centralized logging with level filtering and message filtering for the Tempest HomeKit application.

## Features

- **Level-based Filtering**: Control verbosity with `debug`, `info`, `warn`/`warning`, and `error` levels
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
logger.Warn("Important adjustment made: %s", adjustment)
logger.Error("Error occurred: %v", err)
logger.Alarm("Critical weather condition detected: %s", condition)
```

### Log Levels

The logger uses a hierarchical level system where each level includes all higher priority levels:

- **`debug`**: Shows all messages (DEBUG, INFO, WARN, ERROR, ALARM)
- **`info`**: Shows INFO, WARN, ERROR, and ALARM messages
- **`warn`** (or **`warning`**): Shows WARN, ERROR, and ALARM messages
- **`error`**: Shows ERROR and ALARM messages (default)

**Note**: ALARM messages always appear regardless of log level - they bypass level filtering.

#### When to Use Each Level

- **`Debug()`**: Detailed technical information for troubleshooting (verbose)
- **`Info()`**: General operational messages about normal system behavior
- **`Warn()`**: Important automatic adjustments or non-critical issues that users should be aware of
- **`Error()`**: Critical errors that require attention or prevent normal operation
- **`Alarm()`**: Weather alarm notifications (always visible, bypasses log level filtering)

### Message Filtering

Filter logs to show only messages containing a specific string (case-insensitive):

```go
// Show only UDP-related messages
logger.SetLogFilter("udp")

// Now only messages containing "udp" (case-insensitive) will be output
logger.Info("UDP packet received") //  Shown
logger.Info("API call completed") //  Hidden
logger.Debug("Processing UDP data") //  Shown
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
 return true // No filter, show all
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

#### `Alarm(format string, v ...interface{})`
Logs an alarm notification (always shown, bypasses log level filtering)

**Note**: The `Alarm()` function is specifically designed for weather alarm notifications and always outputs regardless of the configured log level. This ensures critical alarm messages are never suppressed. See [ALARM_LOGGING.md](../../pkg/alarm/docs/ALARM_LOGGING.md) for details.

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
