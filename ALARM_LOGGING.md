# Alarm Logging Behavior

## Overview

Alarm notifications in the Tempest HomeKit system are treated as **critical events** that bypass the standard log level filtering. This ensures that alarm messages are always visible regardless of the configured log level.

## Design Decision

Alarms are not regular log messages - they are important notifications about weather conditions that require user attention. Therefore:

- ✅ **Alarms always display** regardless of log level (debug, info, warn, error)
- ✅ **Alarms use a distinct prefix** (`🚨 ALARM:`) to make them easy to identify
- ✅ **Alarms are not filtered** by the log filter setting
- ✅ **Alarms appear in console output** even with `--loglevel error` or `--loglevel warning`

## Implementation

### Logger Function

A dedicated `Alarm()` function was added to the logger package that bypasses log level checks:

```go
// Alarm always prints alarm notifications, bypassing log level filtering
// Alarms are critical events that should always be visible
func Alarm(format string, v ...interface{}) {
    message := fmt.Sprintf(format, v...)
    if shouldLog(message) {
        log.Printf("🚨 ALARM: %s", message)
    }
}
```

### Console Notifier

The console notifier uses `logger.Alarm()` instead of `logger.Info()`:

```go
func (n *ConsoleNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
    message := expandTemplate(channel.Template, alarm, obs, stationName)
    logger.Alarm("%s", message)
    return nil
}
```

## Examples

### With Warning Log Level

```bash
$ ./tempest-homekit-go --loglevel warning --alarms @tempest-alarms.json

# Only ALARM and ERROR messages appear (INFO/DEBUG suppressed)
2025/10/10 20:51:30 🚨 ALARM: 🚨 ALARM: Wind Change
Station: Chino Hills
Time: 2025-10-10 20:50:51 PDT
Description: Let me know when the wind changes
Wind speed: 0.2
Last Wind Speed: 0.1
```

### With Debug Log Level

```bash
$ ./tempest-homekit-go --loglevel debug --alarms @tempest-alarms.json

# All messages appear including ALARM
2025/10/10 20:52:19 INFO: Starting service...
2025/10/10 20:52:19 DEBUG: Fetching stations...
2025/10/10 20:52:45 🚨 ALARM: 🚨 ALARM: Wind Change
Station: Chino Hills
Time: 2025-10-10 20:50:51 PDT
...
```

### With Error Log Level

Even with the most restrictive log level, alarms still appear:

```bash
$ ./tempest-homekit-go --loglevel error --alarms @tempest-alarms.json

# Only ALARM and ERROR messages appear
2025/10/10 20:51:30 🚨 ALARM: 🚨 ALARM: Wind Change
Station: Chino Hills
...
```

## Other Notification Channels

This behavior applies specifically to the **console** notification channel. Other channels have their own output mechanisms:

- **syslog**: Uses system syslog with configured priority
- **oslog**: Uses macOS unified logging system (os_log API)
- **eventlog**: Uses Windows Event Log (Windows only)
- **email**: Sends via SMTP
- **sms**: Sends via SMS gateway

These channels are not affected by the application's log level setting as they have separate output destinations.

## Testing

To test that alarms bypass log filtering:

```bash
# Start with warning level (suppresses INFO/DEBUG)
./tempest-homekit-go --loglevel warning --alarms @tempest-alarms.json

# Verify alarm status via API
curl http://localhost:8080/api/alarm-status

# Wait for an alarm to trigger (e.g., Wind Change with 10s cooldown)
# Alarm messages should appear in output even though INFO/DEBUG are suppressed
```

## Historical Context

**Previous Behavior**: Alarm console notifications used `logger.Info()`, which meant they were hidden when running with `--loglevel warning` or `--loglevel error`.

**Issue**: Users running in production with restricted log levels would miss critical alarm notifications.

**Solution**: Created dedicated `logger.Alarm()` function that always outputs, ensuring alarms are never suppressed by log level filtering.

## Related Files

- `pkg/logger/logger.go` - Defines the `Alarm()` function
- `pkg/alarm/notifiers.go` - Console notifier implementation
- `tempest-alarms.json` - Alarm configuration file
- `test-alarm-console.sh` - Test script for alarm visibility

## See Also

- [Alarm System Documentation](pkg/alarm/README.md)
- [Logger Documentation](pkg/logger/README.md)
- [OSLog Notifier Documentation](OSLOG_NOTIFIER.md)
