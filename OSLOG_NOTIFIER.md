# macOS OSLog Notification Channel

## Overview

The `oslog` notification channel type provides native integration with macOS's unified logging system. Unlike the traditional `syslog` channel (which uses the legacy BSD syslog API), `oslog` writes directly to the macOS unified logging system, making alarm messages easily accessible through Console.app and the `log` command.

## Implementation

The OSLog notifier integrates with macOS unified logging system, providing a native way to view and filter alarm notifications on macOS systems.

- **Subsystem**: `com.bci.tempest-homekit`
- **Category**: `alarm`
- **Build tags**: Uses `//go:build darwin` to compile only on macOS

## Configuration

To use the oslog channel, add it to your alarm configuration:

```json
{
  "alarms": [
    {
      "name": "Wind Change",
      "description": "Alert when wind speed changes",
      "enabled": true,
      "condition": "*wind_speed",
      "cooldown": 60,
      "channels": [
        {
          "type": "oslog",
          "template": "🚨 ALARM: {{alarm_name}}\nStation: {{station}}\nWind: {{wind_speed}} m/s (was {{last_wind_speed}})"
        }
      ]
    }
  ]
}
```

## Viewing OSLog Messages

### Using Console.app (GUI)
1. Open **Console.app** (Applications > Utilities)
2. In the search bar, filter by:
   - **Subsystem**: `com.behr.tempest-homekit`
   - **Category**: `alarm`
3. Messages will appear in real-time

### Using Terminal (log command)

**Stream messages in real-time:**
```bash
# Stream real-time logs
log stream --predicate 'subsystem == "com.bci.tempest-homekit"'

**Query recent messages:**
```bash
```bash
# Show last hour of logs
log show --predicate 'subsystem == "com.bci.tempest-homekit"' --last 1h --info

# Show last 10 minutes in syslog format
log show --predicate 'subsystem == "com.bci.tempest-homekit"' --last 10m --info --style syslog

# Filter by category
log show --predicate 'subsystem == "com.bci.tempest-homekit" AND category == "alarm"' --last 1h
```

**Export to file:**
```bash
# Export to file
log show --predicate 'subsystem == "com.bci.tempest-homekit"' --last 24h > alarms.log
```
```

## Comparison: syslog vs oslog on macOS

| Feature | syslog | oslog |
|---------|--------|-------|
| **Platform** | Linux, macOS, Unix | macOS only |
| **API** | BSD syslog (legacy) | macOS unified logging |
| **Visibility** | Not visible in unified log on macOS | Fully integrated with Console.app and `log` command |
| **Persistence** | Limited on modern macOS | Persistent in system log database |
| **Querying** | Difficult on macOS | Easy with `log` command predicates |
| **Configuration** | Can specify priority, remote server | No additional config needed |

## Recommendations

- **macOS users**: Use `oslog` for local logging
- **Linux users**: Use `syslog` (oslog not available)
- **Remote logging**: Use `syslog` with network configuration
- **Cross-platform**: Use both channels or `console` channel

## Example: Using Both syslog and oslog

```json
{
  "channels": [
    {
      "type": "console",
      "template": "🚨 {{alarm_name}}: {{alarm_description}}"
    },
    {
      "type": "oslog",
      "template": "🚨 ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}"
    },
    {
      "type": "syslog",
      "template": "ALARM: {{alarm_name}} - {{alarm_description}}"
    }
  ]
}
```

This configuration:
- Shows immediate feedback in console output
- Logs to macOS unified logging (oslog) for macOS users
- Logs to traditional syslog for Linux deployment or remote syslog servers

## Testing

Test the oslog integration:

```bash
# Run with debug logging
./tempest-homekit-go --loglevel debug --alarms @tempest-alarms.json

# In another terminal, stream the logs
log stream --predicate 'subsystem == "com.bci.tempest-homekit"'

# Trigger an alarm and watch it appear in both terminals
```

## Technical Details

### CGO Implementation
The oslog notifier uses CGO to call the native macOS `os_log_create` and `os_log_with_type` functions:

```c
#include <os/log.h>

void log_message(const char *subsystem, const char *category, const char *message) {
    os_log_t log = os_log_create(subsystem, category);
    os_log_with_type(log, OS_LOG_TYPE_DEFAULT, "%{public}s", message);
}
```

### Build Requirements
- macOS SDK with `os/log.h` header
- CGO enabled (default for Go on macOS)
- Framework linking: `-framework Foundation`

### Platform Compatibility
On non-macOS platforms, the oslog notifier returns an error:
```
oslog notification type is only supported on macOS
```

## Related Documentation
- [Alarm System Overview](README.md)
- [Template Variables](../../ALARM_EDITOR_VARIABLES.md)
- [Apple's Unified Logging Documentation](https://developer.apple.com/documentation/os/logging)
